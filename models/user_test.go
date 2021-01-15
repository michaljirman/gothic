package models

import (
	"testing"

	"github.com/jrapoport/gothic/conf"
	"github.com/jrapoport/gothic/storage"
	"github.com/jrapoport/gothic/storage/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const modelsTestConfig = "../env/test.env"

type UserTestSuite struct {
	suite.Suite
	db *storage.Connection
}

func (ts *UserTestSuite) SetupTest() {
	//storage.TruncateAll(ts.db)
}

func (ts *UserTestSuite) TearDownTest() {
	storage.TruncateAll(ts.db)
}

func TestUser(t *testing.T) {
	config, err := conf.LoadConfiguration(modelsTestConfig)
	require.NoError(t, err)

	conn, err := test.SetupDBConnection(t, config)
	require.NoError(t, err)

	ts := &UserTestSuite{
		db: conn,
	}

	suite.Run(t, ts)
}

func (ts *UserTestSuite) TestUpdateAppMetadata() {
	u, err := NewUser("", "", nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), u.UpdateAppMetaData(ts.db, make(map[string]interface{})))

	require.NotNil(ts.T(), u.AppMetaData)

	require.NoError(ts.T(), u.UpdateAppMetaData(ts.db, map[string]interface{}{
		"foo": "bar",
	}))

	require.Equal(ts.T(), "bar", u.AppMetaData["foo"])
	require.NoError(ts.T(), u.UpdateAppMetaData(ts.db, map[string]interface{}{
		"foo": nil,
	}))
	require.Len(ts.T(), u.AppMetaData, 0)
	require.Equal(ts.T(), nil, u.AppMetaData["foo"])
}

func (ts *UserTestSuite) TestUpdateUserMetadata() {
	u, err := NewUser("", "", nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), u.UpdateUserMetaData(ts.db, make(map[string]interface{})))

	require.NotNil(ts.T(), u.UserMetaData)

	require.NoError(ts.T(), u.UpdateUserMetaData(ts.db, map[string]interface{}{
		"foo": "bar",
	}))

	require.Equal(ts.T(), "bar", u.UserMetaData["foo"])
	require.NoError(ts.T(), u.UpdateUserMetaData(ts.db, map[string]interface{}{
		"foo": nil,
	}))
	require.Len(ts.T(), u.UserMetaData, 0)
	require.Equal(ts.T(), nil, u.UserMetaData["foo"])
}

func (ts *UserTestSuite) TestFindUserByConfirmationToken() {
	u := ts.createUser()

	n, err := FindUserByConfirmationToken(ts.db, u.ConfirmationToken)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), u.ID, n.ID)
}

func (ts *UserTestSuite) TestFindUserByEmail() {
	u := ts.createUser()

	n, err := FindUserByEmail(ts.db, u.Email)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), u.ID, n.ID)

	_, err = FindUserByEmail(ts.db, u.Email+"m")
	require.EqualError(ts.T(), err, UserNotFoundError{}.Error())
}

func (ts *UserTestSuite) TestFindUsersInAudience() {
	_ = ts.createUser()
	n, err := FindUsers(ts.db, nil, nil, "")
	require.NoError(ts.T(), err)
	require.Len(ts.T(), n, 1)

	p := Pagination{
		Page:    1,
		PerPage: 50,
	}
	n, err = FindUsers(ts.db, &p, nil, "")
	require.NoError(ts.T(), err)
	require.Len(ts.T(), n, 1)
	assert.Equal(ts.T(), uint64(1), p.Count)

	sp := &SortParams{
		Fields: []SortField{
			SortField{Name: "created_at", Dir: Descending},
		},
	}
	n, err = FindUsers(ts.db, nil, sp, "")
	require.NoError(ts.T(), err)
	require.Len(ts.T(), n, 1)
}

func (ts *UserTestSuite) TestFindUserByID() {
	u := ts.createUser()

	n, err := FindUserByID(ts.db, u.ID)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), u.ID, n.ID)
}

func (ts *UserTestSuite) TestFindUserByRecoveryToken() {
	u := ts.createUser()
	u.RecoveryToken = "asdf"

	err := ts.db.Save(u).Error
	require.NoError(ts.T(), err)

	n, err := FindUserByRecoveryToken(ts.db, u.RecoveryToken)
	require.NoError(ts.T(), err)

	require.Equal(ts.T(), u.ID, n.ID)
}

func (ts *UserTestSuite) TestFindUserWithRefreshToken() {
	u := ts.createUser()
	r, err := GrantAuthenticatedUser(ts.db, u)
	require.NoError(ts.T(), err)

	n, nr, err := FindUserWithRefreshToken(ts.db, r.Token)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), r.ID, nr.ID)
	require.Equal(ts.T(), u.ID, n.ID)
}

func (ts *UserTestSuite) TestIsDuplicatedEmail() {
	_ = ts.createUserWithEmail("test.user@gothic.com")

	e, err := IsDuplicatedEmail(ts.db, "test.user@gothic.com")
	require.NoError(ts.T(), err)
	require.True(ts.T(), e, "expected email to be duplicated")

	e, err = IsDuplicatedEmail(ts.db, "testuser@gothic.com")
	require.NoError(ts.T(), err)
	require.False(ts.T(), e, "expected email to not be duplicated")
}

func (ts *UserTestSuite) createUser() *User {
	return ts.createUserWithEmail("testuser@gothic.com")
}

func (ts *UserTestSuite) createUserWithEmail(email string) *User {
	user, err := NewUser(email, "secret", nil)
	require.NoError(ts.T(), err)

	err = ts.db.Create(user).Error
	require.NoError(ts.T(), err)

	return user
}
