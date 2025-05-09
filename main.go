package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// get reference to the local source code directory
	src := client.Host().Directory("./node-app")

	// setup the build container
	packcli := client.Container().From("paketobuildpacks/builder-jammy-full:latest")

	// Mount the local source code directory into the container
	packcli = packcli.WithMountedDirectory("/tmp/src", src)

	// Make a temporary directory in the container to copy the source code into because of permissions
	packcli = packcli.WithExec([]string{"mkdir", "/tmp/src1"})
	packcli = packcli.WithExec([]string{"cp", "-r", "/tmp/src/.", "/tmp/src1/"})

	//set the working directory to the temporary directory
	packcli = packcli.WithWorkdir("/tmp/src1")

	packcli = packcli.WithExec([]string{"ls", "-al"})
	packcli = packcli.WithExec([]string{"pwd"})

	// packcli = packcli.WithExec([]string{"ls", "-al", "/tmp/src"})

	packcli = packcli.WithExec([]string{"bash", "-c", fmt.Sprintf("CNB_PLATFORM_API=0.14 /cnb/lifecycle/creator -app=. %s", "demo-node-app:10m")})

	// define the application build command
	// packcli = packcli.WithExec([]string{
	// 	"pack", "build", "ttl.sh/demo-node-app:2h",
	// 	"--path", "node-app",
	// 	"--builder", "heroku/builder:24",
	// 	// "--buildpack", "paketo-buildpacks/nodejs",
	// 	// "--cache", "type=build;format=bind;source=/tmp/build-cache;type=launch;format=bind;source=/tmp/build-cache",
	// 	"--env", "BP_DISABLE_SBOM=true",
	// 	// "--volume", "/tmp/:/tmp/build-cache/",
	// 	// "--clear-cache",
	// 	// "--platform", "linux/arm64",
	// 	"--verbose",
	// 	"--publish",
	// 	// "--creation-time", "now",
	// 	// any other --env flags here too
	// })
	packcli = packcli.WithExec([]string{"echo", "after pack"})

	_, err = packcli.Stdout(ctx)
	if err != nil {
		return err
	}

	// fmt.Println(out)

	return nil
}
