package impl

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imulab/homelab/proxmox/login/api"
	"github.com/imulab/homelab/proxmox/ticket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"sync"
)

var (
	oneDefaultService      sync.Once
	defaultServiceInstance *defaultService
)

func DefaultService() api.Service {
	oneDefaultService.Do(func() {
		defaultServiceInstance = newDefaultService()
	})
	return defaultServiceInstance
}

func newDefaultService() *defaultService {
	return &defaultService{
		ticketService: ticket.DefaultService(),
		loginApi:      "%s/api2/json/access/ticket",
	}
}

type defaultService struct {
	ticketService ticket.Service
	loginApi      string
}

func (s *defaultService) Login(request *api.Request) (*api.Response, error) {
	if cached, err := s.ticketService.Get(); err != nil && request.Force {
		log.WithFields(log.Fields{
			"event": "login_success",
			"login": true,
			"cache": map[string]interface{}{
				"hit":     true,
				"storage": s.ticketService.DefaultStorage(),
			},
		}).Debug("reuse cached session.")

		return &api.Response{
			ApiServer:      cached.ApiServer,
			Username:       cached.Username,
			SessionStorage: s.ticketService.DefaultStorage(),
		}, nil
	}

	return s.doLogin(request)
}

func (s *defaultService) doLogin(request *api.Request) (*api.Response, error) {
	var session = new(ticket.Session)

	err := s.postRequest(request, func(status int, response *http.Response) error {
		switch status {
		case http.StatusOK:
			data := make(map[string]interface{})
			if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
				return err
			}

			session.Username = data["data"].(map[string]interface{})["username"].(string)
			session.Ticket = data["data"].(map[string]interface{})["ticket"].(string)
			session.CSRFToken = data["data"].(map[string]interface{})["CSRFPreventionToken"].(string)
			session.ApiServer = request.ApiServer
			return nil

		case http.StatusUnauthorized:
			return errors.New("unauthorized")

		default:
			log.WithFields(log.Fields{
				"event":  "http_debug",
				"status": status,
			}).Debug("http post received non-200/non-401 status.")
			return errors.New("bad_request")
		}
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"event": "login_fail",
		}).Error("failed to login and/or acquire ticket.")
		return nil, err
	}

	if session == nil {
		panic("session should not be nil here.")
	}

	if err = s.ticketService.Save(session); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"event": "ticket_failed",
		}).Error("failed to save session.")
		return nil, errors.New("ticket_error")
	}

	return &api.Response{
		ApiServer:      session.ApiServer,
		Username:       session.Username,
		SessionStorage: s.ticketService.DefaultStorage(),
	}, nil
}

func (s *defaultService) postRequest(request *api.Request, responseLogic func(status int, response *http.Response) error) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	form := url.Values{}
	form.Add("username", request.Username)
	form.Add("password", request.Password)
	form.Add("realm", request.Realm)

	loginUrl := s.loginURL(request.ApiServer)
	resp, err := client.PostForm(loginUrl, form)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"event": "http_fail",
			"url":   loginUrl,
		}).Error("failed to post request.")
		return err
	}
	defer resp.Body.Close()

	if err = responseLogic(resp.StatusCode, resp); err != nil {
		return err
	}

	return nil
}

func (s *defaultService) loginURL(serverURL string) string {
	return fmt.Sprintf(s.loginApi, serverURL)
}
