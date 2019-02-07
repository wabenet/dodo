package image

import (
	buildkit "github.com/moby/buildkit/session"
	"golang.org/x/net/context"
)

type fakeSession struct{}

func (session *fakeSession) ID() string {
	return "MySessionID"
}

func (session *fakeSession) Allow(buildkit.Attachable) {
	return
}

func (session *fakeSession) Run(context.Context, buildkit.Dialer) error {
	return nil
}

func (session *fakeSession) Close() error {
	return nil
}
