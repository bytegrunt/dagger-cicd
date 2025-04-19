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
	src := client.Host().Directory(".")
	packcli := client.Container().From("docker:28.1-cli").WithUnixSocket("/var/run/docker.sock", client.Host().UnixSocket("unix:///var/run/docker.sock")) // Alpine-based, has Docker CLI
	packcli = packcli.
		WithExec([]string{"apk", "add", "--no-cache", "curl", "tar"}).
		WithExec([]string{"sh", "-c", `curl -sSL https://github.com/buildpacks/pack/releases/download/v0.37.0/pack-v0.37.0-linux.tgz | tar -xz -C /usr/local/bin`})

	// packcli := client.Container().From("ttl.sh/my_pack_cont:1h").WithUnixSocket("/var/run/docker.sock", client.Host().UnixSocket("unix:///var/run/docker.sock"))

	// mount cloned repository into `golang` image
	packcli = packcli.WithDirectory("./node-app", src).WithWorkdir("./node-app")
	// packcli = packcli.WithExec([]string{"apt", "get", "docker"})

	// packcli = packcli.WithExec([]string{"ls", "-l", "/var/run/docker.sock"})
	packcli = packcli.WithExec([]string{"echo", "before pack"})
	// packcli = packcli.WithExec([]string{"ls", "-l", "/var/run/docker.sock"})
	// packcli = packcli.WithUser("root")
	// packcli = packcli.WithExec([]string{"docker", "info"})
	// packcli = packcli.WithExec([]string{"docker", "ps"})
	// packcli = packcli.WithExec([]string{"docker", "run", "--rm", "alpine", "echo", "hello-from-dagger"})
	// packcli = packcli.WithExec([]string{
	// 	"docker", "tag", "alpine", "ttl.sh/dagger-test-push:1h",
	// })
	// packcli = packcli.WithExec([]string{
	// 	"docker", "push", "ttl.sh/dagger-test-push:1h",
	// })

	// packcli = packcli.WithExec([]string{
	// 	"sh", "-c",
	// 	"pack build ttl.sh/demo-node-app:2h --path node-app --builder heroku/builder:24 --env BP_DISABLE_SBOM=true --platform linux/arm64 --verbose --publish && echo 'Pack finished cleanly'",
	// })

	packcli = packcli.WithExec([]string{
		"sh", "-c",
		"nohup pack build ttl.sh/demo-node-app:2h " +
			"--path node-app " +
			"--builder heroku/builder:24 " +
			"--env BP_DISABLE_SBOM=true " +
			"--platform linux/arm64 " +
			"--verbose " +
			"--publish > /tmp/pack.log 2>&1 && touch /tmp/pack.done &",
	})

	doneFile := packcli.File("/tmp/pack.done")

	for i := 0; i < 40; i++ { // wait max ~2min
		_, err := doneFile.Contents(ctx)
		if err == nil {
			fmt.Println("✅ Pack finished.")
			break
		}
		fmt.Println("⏳ Waiting for pack to finish...")
		time.Sleep(3 * time.Second)
	}

	logFile := packcli.File("/tmp/pack.log")
	contents, err := logFile.Contents(ctx)
	if err == nil {
		fmt.Println("📄 Pack output:\n", contents)
	}

	// packcli = packcli.WithExec([]string{
	// 	"sh", "-c",
	// 	"pack build ttl.sh/demo-node-app:2h " +
	// 		"--path node-app " +
	// 		"--builder heroku/builder:24 " +
	// 		"--env BP_DISABLE_SBOM=true " +
	// 		"--platform linux/arm64 " +
	// 		"--verbose " +
	// 		"--publish && touch /tmp/pack.done",
	// })

	// doneFile := packcli.File("/tmp/pack.done")
	// _, err = doneFile.Contents(ctx)
	// if err != nil {
	// 	fmt.Println("❌ Pack did not complete or signal file not found:", err)
	// } else {
	// 	fmt.Println("✅ Pack completed successfully (via done file)")
	// }

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
