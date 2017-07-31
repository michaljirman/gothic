package api

import (
	"encoding/json"
	"net/http"

	"github.com/netlify/gotrue/models"
)

// UserUpdateParams parameters for updating a user
type UserUpdateParams struct {
	Email            string                 `json:"email"`
	Password         string                 `json:"password"`
	EmailChangeToken string                 `json:"email_change_token"`
	Data             map[string]interface{} `json:"data"`
	AppData          map[string]interface{} `json:"app_metadata,omitempty"`
}

// UserGet returns a user
func (a *API) UserGet(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	token := getToken(ctx)

	id, ok := token.Claims["id"].(string)
	if !ok {
		return badRequestError("Could not read User ID claim")
	}

	tokenAud, ok := token.Claims["aud"].(string)
	if !ok {
		return badRequestError("Could not read User Aud claim")
	}

	aud := a.requestAud(ctx, r)
	if aud != tokenAud {
		return badRequestError("Token audience doesn't match request audience")
	}

	user, err := a.db.FindUserByID(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return notFoundError(err.Error())
		}
		return internalServerError("Database error finding user").WithInternalError(err)
	}

	return sendJSON(w, http.StatusOK, user)
}

// UserUpdate updates fields on a user
func (a *API) UserUpdate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	token := getToken(ctx)

	params := &UserUpdateParams{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)
	if err != nil {
		return badRequestError("Could not read User Update params: %v", err)
	}

	id, ok := token.Claims["id"].(string)
	if !ok {
		return badRequestError("Could not read User ID claim")
	}

	user, err := a.db.FindUserByID(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return notFoundError(err.Error())
		}
		return internalServerError("Database error finding user").WithInternalError(err)
	}

	var sendChangeEmailVerification bool
	if params.Email != "" {
		if err = a.mailer.ValidateEmail(params.Email); err == nil {
			exists, err := a.db.IsDuplicatedEmail(params.Email, user.Aud)
			if err != nil {
				return internalServerError("Database error checking email").WithInternalError(err)
			}

			if exists {
				return unprocessableEntityError("Email address already registered by another user")
			}

			user.GenerateEmailChange(params.Email)
			sendChangeEmailVerification = true
		} else {
			return unprocessableEntityError("Unable to verify new email address: " + err.Error())
		}
	}

	log := getLogEntry(r)
	log.Debugf("Checking params for token %v", params)

	if params.EmailChangeToken != "" {
		log.Debugf("Got change token %v", params.EmailChangeToken)

		if params.EmailChangeToken != user.EmailChangeToken {
			return unauthorizedError("Email Change Token didn't match token on file")
		}

		user.ConfirmEmailChange()
	}

	if params.Password != "" {
		if err = user.EncryptPassword(params.Password); err != nil {
			return internalServerError("Error during password encryption").WithInternalError(err)
		}
	}

	if params.Data != nil {
		user.UpdateUserMetaData(params.Data)
	}

	if params.AppData != nil {
		if a.isAdmin(user, a.config.JWT.Aud) {
			return unauthorizedError("Updating app_metadata requires admin privileges")
		}

		user.UpdateAppMetaData(params.AppData)
	}

	if err := a.db.UpdateUser(user); err != nil {
		return internalServerError("Database error updating user").WithInternalError(err)
	}

	if sendChangeEmailVerification {
		if err = a.mailer.EmailChangeMail(user); err != nil {
			log := getLogEntry(r)
			log.WithError(err).Error("Error sending change email")
		}
	}

	return sendJSON(w, http.StatusOK, user)
}
