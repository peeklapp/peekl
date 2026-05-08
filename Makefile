test-podman:
	podman rmi peekl-test:latest || true
	podman build -t peekl-test:latest -f tests.Dockerfile
	podman run --rm --user=root peekl-test:latest
