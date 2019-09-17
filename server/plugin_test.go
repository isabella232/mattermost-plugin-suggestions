package main

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockReader(filename string) ([]byte, error) {
	return []byte{0}, nil
}

func TestSetupBot(t *testing.T) {
	t.Run("EnsureBot error", func(t *testing.T) {
		plugin := &Plugin{}
		helpers := &plugintest.Helpers{}
		helpers.On("EnsureBot", mock.Anything).Return("", errors.New(""))
		plugin.SetHelpers(helpers)
		defer helpers.AssertExpectations(t)
		err := plugin.setupBot(nil)
		assert.NotNil(t, err)
	})

	t.Run("GetBundlePath error", func(t *testing.T) {
		plugin := &Plugin{}
		helpers := &plugintest.Helpers{}
		helpers.On("EnsureBot", mock.Anything).Return("bot", nil)
		plugin.SetHelpers(helpers)
		defer helpers.AssertExpectations(t)
		api := &plugintest.API{}
		api.On("GetBundlePath").Return("", errors.New(""))
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		err := plugin.setupBot(nil)
		assert.NotNil(t, err)
	})

	t.Run("wrong file path", func(t *testing.T) {
		plugin := &Plugin{}
		helpers := &plugintest.Helpers{}
		helpers.On("EnsureBot", mock.Anything).Return("bot", nil)
		plugin.SetHelpers(helpers)
		defer helpers.AssertExpectations(t)
		api := &plugintest.API{}
		api.On("GetBundlePath").Return("some path", nil)
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		err := plugin.setupBot(ioutil.ReadFile)
		assert.NotNil(t, err)
	})

	t.Run("SetProfileImage error", func(t *testing.T) {
		plugin := &Plugin{}
		helpers := &plugintest.Helpers{}
		helpers.On("EnsureBot", mock.Anything).Return("bot", nil)
		plugin.SetHelpers(helpers)
		defer helpers.AssertExpectations(t)
		api := &plugintest.API{}
		api.On("GetBundlePath").Return("some path", nil)
		api.On("SetProfileImage", mock.Anything, mock.Anything).Return(model.NewAppError("", "", nil, "", 404))
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		err := plugin.setupBot(mockReader)
		assert.NotNil(t, err)
	})

	t.Run("no error", func(t *testing.T) {
		plugin := &Plugin{}
		helpers := &plugintest.Helpers{}
		helpers.On("EnsureBot", mock.Anything).Return("bot", nil)
		plugin.SetHelpers(helpers)
		defer helpers.AssertExpectations(t)
		api := &plugintest.API{}
		api.On("GetBundlePath").Return("some path", nil)
		api.On("SetProfileImage", mock.Anything, mock.Anything).Return(nil)
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		err := plugin.setupBot(mockReader)
		assert.Nil(t, err)
	})
}

func TestOnActivate(t *testing.T) {
	t.Run("EnsureBot error", func(t *testing.T) {
		plugin := &Plugin{}
		api := &plugintest.API{}
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		helpers := &plugintest.Helpers{}
		plugin.SetHelpers(helpers)
		defer helpers.AssertExpectations(t)

		api.On("RegisterCommand", mock.Anything).Return(nil)
		// helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return(nil)
		helpers.On("EnsureBot", mock.Anything).Return("", errors.New(""))

		err := plugin.OnActivate()
		assert.NotNil(t, err)
	})
}
