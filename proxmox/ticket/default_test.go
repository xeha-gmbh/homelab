package ticket

import (
	"encoding/json"
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

type DefaultServiceTestSuite struct {
	suite.Suite
	service       Service
	sampleSession string
}

func (suite *DefaultServiceTestSuite) SetupSuite() {
	suite.service = newDefaultService()
	suite.sampleSession = dedent.Dedent(`
	{
		"username": "root@pam",
		"ticket": "foo",
		"csrf_token": "bar",
		"api_server": "http://localhost:8006"
	}`)
}

func (suite *DefaultServiceTestSuite) SetupTest() {
	if _, err := os.Stat(suite.service.DefaultStorage()); !os.IsNotExist(err) {
		os.Rename(suite.service.DefaultStorage(), suite.service.DefaultStorage()+".bak")
	}
}

func (suite *DefaultServiceTestSuite) TearDownTest() {
	if _, err := os.Stat(suite.service.DefaultStorage() + ".bak"); !os.IsNotExist(err) {
		os.Rename(suite.service.DefaultStorage()+".bak", suite.service.DefaultStorage())
	}
}

func (suite *DefaultServiceTestSuite) TestGet() {
	var (
		writeSample = func() error {
			if err := ioutil.WriteFile(suite.service.DefaultStorage(), []byte(suite.sampleSession), 0644); err != nil {
				return err
			}
			return nil
		}
		removeSample = func() error {
			return os.Remove(suite.service.DefaultStorage())
		}
	)

	suite.Require().Nil(writeSample())
	defer func() {
		suite.Require().Nil(removeSample())
	}()

	session, err := suite.service.Get()
	suite.Assert().Nil(err, "Failed get by default service.")

	suite.Assert().Equal("root@pam", session.Username)
	suite.Assert().Equal("foo", session.Ticket)
	suite.Assert().Equal("bar", session.CSRFToken)
	suite.Assert().Equal("http://localhost:8006", session.ApiServer)
}

func (suite *DefaultServiceTestSuite) TestWrite() {
	var (
		session           = new(Session)
		deserializeSample = func() error {
			return json.Unmarshal([]byte(suite.sampleSession), session)
		}
		removeSample = func() error {
			return os.Remove(suite.service.DefaultStorage())
		}
	)

	suite.Require().Nil(deserializeSample())
	defer func() {
		suite.Require().Nil(removeSample())
	}()

	suite.Assert().Nil(suite.service.Save(session), "Failed save by default service.")

	f, err := os.Open(suite.service.DefaultStorage())
	suite.Assert().Nil(err)

	mirror := new(Session)
	err = json.NewDecoder(f).Decode(mirror)
	suite.Assert().Nil(err)

	suite.Assert().Equal(session.Username, mirror.Username)
	suite.Assert().Equal(session.Ticket, mirror.Ticket)
	suite.Assert().Equal(session.CSRFToken, mirror.CSRFToken)
	suite.Assert().Equal(session.ApiServer, mirror.ApiServer)
}

func TestDefaultService(t *testing.T) {
	suite.Run(t, new(DefaultServiceTestSuite))
}
