package main

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// a job for pre-calculating channel recommendations for users.
	preCalcJob    *cron.Cron
	preCalcPeriod string
	botUserID     string
}

type readFile func(filename string) ([]byte, error)

func (p *Plugin) setupBot(reader readFile) error {
	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    "suggestions",
		DisplayName: "Suggestions",
		Description: "Created by the Suggestions plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure suggestions bot")
	}
	p.botUserID = botID
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	profileImage, err := reader(filepath.Join(bundlePath, "assets", "profile.jpeg"))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	appErr := p.API.SetProfileImage(botID, profileImage)
	if appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}
	return nil
}

func (p *Plugin) startPrecalcJob() error {
	config := p.getConfiguration()
	p.preCalcPeriod = config.PreCalculationPeriod
	c := cron.New()
	_, err := c.AddFunc(p.preCalcPeriod, func() {
		err := p.RunOnSingleNode(func() {
			if appErr := p.preCalculateRecommendations(); appErr != nil {
				p.API.LogError("Can't calculate recommendations", "error", appErr)
			}
		})
		if err != nil {
			p.API.LogError("Can't run on single node", "error", err)
		}
	})
	if err != nil {
		return err
	}
	c.Start()
	p.preCalcJob = c
	return nil
}

// OnActivate will be run on plugin activation.
func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(getCommand())
	err := p.setupBot(ioutil.ReadFile)
	if err != nil {
		return err
	}
	if err = p.startPrecalcJob(); err != nil {
		return errors.Wrap(err, "Can't start precalc job")
	}
	go func() { //precalculate at once
		err := p.RunOnSingleNode(func() {
			if appErr := p.preCalculateRecommendations(); appErr != nil {
				p.API.LogError("Can't calculate recommendations", "error", appErr)
			}
		})
		if err != nil {
			p.API.LogError("Can't run on single node", "error", err)
		}
	}()
	return nil
}

// OnDeactivate will be run on plugin deactivation.
func (p *Plugin) OnDeactivate() error {
	if p.preCalcJob != nil {
		p.preCalcJob.Stop()
	}
	return nil
}
