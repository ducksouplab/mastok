package front

import (
	"log"

	"github.com/ducksouplab/mastok/env"
	"github.com/evanw/esbuild/pkg/api"
)

// API

func Build() {
	devEnv := env.Mode == "DEV"
	logPlugin := api.Plugin{
		Name: "log",
		Setup: func(build api.PluginBuild) {
			build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				if len(result.Errors) > 0 {
					for _, msg := range result.Errors {
						log.Println("[JS] build error: " + msg.Text)
					}
				} else {
					if len(result.Warnings) > 0 {
						for _, msg := range result.Warnings {
							log.Println("[JS] build warning: " + msg.Text)
						}
					} else {
						log.Println("[JS] build success")
					}
				}
				return api.OnEndResult{}, nil
			})
		},
	}

	buildOptions := api.BuildOptions{
		EntryPoints:       []string{"front/src/join.js", "front/src/supervise.js", "front/src/form.js"},
		Bundle:            true,
		MinifyWhitespace:  !devEnv,
		MinifyIdentifiers: !devEnv,
		MinifySyntax:      !devEnv,
		Engines: []api.Engine{
			{api.EngineChrome, "64"},
			{api.EngineFirefox, "53"},
			{api.EngineSafari, "11"},
			{api.EngineEdge, "79"},
		},
		Outdir:  "front/static/assets/scripts",
		Plugins: []api.Plugin{logPlugin},
		Write:   true,
	}
	if devEnv {
		ctx, err := api.Context(buildOptions)
		if err != nil {
			log.Fatal(err)
		}

		watchErr := ctx.Watch(api.WatchOptions{})
		if watchErr != nil {
			log.Fatal(watchErr)
		}
	} else {
		build := api.Build(buildOptions)

		if len(build.Errors) > 0 {
			log.Fatal("[JS] build error: " + build.Errors[0].Text)
		}
	}
}
