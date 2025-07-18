// Package publish contains the publishing pipe.
package publish

import (
	"fmt"

	"github.com/goreleaser/goreleaser/v2/internal/middleware/errhandler"
	"github.com/goreleaser/goreleaser/v2/internal/middleware/logging"
	"github.com/goreleaser/goreleaser/v2/internal/middleware/skip"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/artifactory"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/aur"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/aursources"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/blob"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/brew"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/cask"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/chocolatey"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/custompublishers"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/docker"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/ko"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/krew"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/milestone"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/nix"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/release"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/scoop"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/sign"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/snapcraft"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/upload"
	"github.com/goreleaser/goreleaser/v2/internal/pipe/winget"
	"github.com/goreleaser/goreleaser/v2/internal/skips"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

// Publisher should be implemented by pipes that want to publish artifacts.
type Publisher interface {
	fmt.Stringer

	// Default sets the configuration defaults
	Publish(ctx *context.Context) error
}

// New publish pipeline.
func New() Pipe {
	return Pipe{
		pipeline: []Publisher{
			blob.Pipe{},
			upload.Pipe{},
			artifactory.Pipe{},
			docker.Pipe{},
			docker.ManifestPipe{},
			ko.Pipe{},
			sign.DockerPipe{},
			snapcraft.Pipe{},
			// This should be one of the last steps
			release.Pipe{},
			// brew et al use the release URL, so, they should be last
			nix.New(),
			winget.Pipe{},
			brew.Pipe{},
			cask.Pipe{},
			aur.Pipe{},
			aursources.Pipe{},
			krew.Pipe{},
			scoop.Pipe{},
			chocolatey.Pipe{},
			milestone.Pipe{},
			custompublishers.Pipe{},
		},
	}
}

// Pipe that publishes artifacts.
type Pipe struct {
	pipeline []Publisher
}

func (Pipe) String() string                 { return "publishing" }
func (Pipe) Skip(ctx *context.Context) bool { return skips.Any(ctx, skips.Publish) }

func (p Pipe) Run(ctx *context.Context) error {
	memo := errhandler.Memo{}
	for _, publisher := range p.pipeline {
		if err := skip.Maybe(
			publisher,
			logging.PadLog(
				publisher.String(),
				errhandler.Handle(publisher.Publish),
			),
		)(ctx); err != nil {
			if ig, ok := publisher.(Continuable); ok && ig.ContinueOnError() && !ctx.FailFast {
				memo.Memorize(fmt.Errorf("%s: %w", publisher.String(), err))
				continue
			}
			return fmt.Errorf("%s: failed to publish artifacts: %w", publisher.String(), err)
		}
	}
	return memo.Error()
}

type Continuable interface {
	ContinueOnError() bool
}
