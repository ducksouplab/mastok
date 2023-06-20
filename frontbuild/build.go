package frontbuild

import (
	"log"

	"github.com/ducksouplab/mastok/env"
	"github.com/evanw/esbuild/pkg/api"
)

// API

func Build() {
	devEnv := env.Mode == "DEV"
	templatesPlugin := api.Plugin{
		Name: "Update templates",
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
						updateTemplates()
						cleanUpAssets()
					}
				}
				return api.OnEndResult{}, nil
			})
		},
	}

	buildOptions := api.BuildOptions{
		EntryPoints: []string{
			"front/src/js/form.js",
			"front/src/js/join.js",
			"front/src/js/supervise.js",
			"front/src/css/consent.css",
			"front/src/css/join.css",
			"front/src/css/main.css",
		},
		EntryNames:        version + "/[dir]/[name]",
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
		Outdir:  "front/static/assets",
		Plugins: []api.Plugin{templatesPlugin},
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
		api.Build(buildOptions)
	}
}
