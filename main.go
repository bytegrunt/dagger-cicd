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
	// get reference to the local project
	src := client.Host().Directory("./node-app")
	packcli := client.Container().From("paketobuildpacks/builder-jammy-base:latest").WithUnixSocket("/var/run/docker.sock", client.Host().UnixSocket("unix:///var/run/docker.sock")) // Alpine-based, has Docker CLI

	
	packcli = packcli.WithDirectory("./src", src) //.WithWorkdir("./node-app")
	
	packcli = packcli.WithExec([]string{"sh","-c","mkdir /tmp/src1; cd /tmp/src; tar -c … | tar -x -C /tmp/src1"})

	packcli = packcli.WithExec([]string{"echo", "before pack"})

	packcli = packcli.WithExec([]string{"bash", "-c", fmt.Sprintf("CNB_PLATFORM_API=0.14 /cnb/lifecycle/creator -app=. %s", "ttl.sh/demo-node-app:30m")})

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

	out, err := packcli.Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Println(out)

	// get reference to build output directory in container
	// output := packcli.Directory("./workspace")

	// // write contents of container build/ directory to the host
	// _, err = output.Export(ctx, "./workspace")
	// if err != nil {
	// 	return err
	// }

	return nil
}
