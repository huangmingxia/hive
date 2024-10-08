package installconfig

import (
	"context"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/validate"
)

type pullSecret struct {
	PullSecret string
}

var _ asset.Asset = (*pullSecret)(nil)

// Dependencies returns no dependencies.
func (a *pullSecret) Dependencies() []asset.Asset {
	return []asset.Asset{}
}

// Generate queries for the pull secret from the user.
func (a *pullSecret) Generate(context.Context, asset.Parents) error {
	if err := survey.Ask([]*survey.Question{
		{
			Prompt: &survey.Password{
				Message: "Pull Secret",
				Help:    "The container registry pull secret for this cluster, as a single line of JSON (e.g. {\"auths\": {...}}).\n\nYou can get this secret from https://console.redhat.com/openshift/install/pull-secret",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				return validate.ImagePullSecret(ans.(string))
			}),
		},
	}, &a.PullSecret); err != nil {
		return errors.Wrap(err, "failed UserInput")
	}
	return nil
}

// Name returns the human-friendly name of the asset.
func (a *pullSecret) Name() string {
	return "Pull Secret"
}
