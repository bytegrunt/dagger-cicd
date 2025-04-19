FROM docker:28.1-cli

RUN apt update && apt install -y \
    curl \
    gnupg \
    ca-certificates \
    && curl -sSL "https://github.com/buildpacks/pack/releases/download/v0.37.0/pack-v0.37.0-linux.tgz" \
    | tar -C /usr/local/bin/ -xzv pack

ENTRYPOINT ["pack"]
