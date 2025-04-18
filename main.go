package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context) error {
	fmt.Println("Building with Dagger")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// get reference to the local project
	src := client.Host().Directory(".")

	packcli := client.Container().From("buildpacksio/pack:latest").WithUnixSocket("/var/run/docker.sock", client.Host().UnixSocket("unix:///var/run/docker.sock"))

	// mount cloned repository into `golang` image
	packcli = packcli.WithDirectory("./node-app", src).WithWorkdir("./node-app")

	// define the application build command
	packcli = packcli.WithExec([]string{
		"pack", "build", "demo-node-app",
		"--path", "node-app",
		"--builder", "heroku/builder:24",
		// "--buildpack", "paketo-buildpacks/nodejs",
		"--env", "BP_DISABLE_SBOM=true",
		"--platform", "linux/arm64",
		// "--creation-time", "now",
		// any other --env flags here too
	})

	// get reference to build output directory in container
	output := packcli.Directory("./workspace")

	// write contents of container build/ directory to the host
	_, err = output.Export(ctx, "./workspace")
	if err != nil {
		return err
	}

	return nil
}
