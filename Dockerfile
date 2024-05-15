# hadolint ignore=DL3007
FROM chainguard/wolfi-base:latest

# Simplest possible form of an image that can utilize hops:
# - Should include Homebrew, but not required
# - Needs a shell

# By default, depends on the artifact from the "build linux" job in the pipeline
ARG HOPS_EXECUTABLE=ci-dist/hops/linux/amd64/bin/hops
# ARG GIT_VERSION=1:2.43.0-1ubuntu1
# ARG CA_CERTIFICATES_VERSION=20230311ubuntu1

# hadolint ignore=DL3008
# RUN \
# 	apt-get update; \
# 	apt-get install -y --no-install-recommends \
# 	# ca-certificates=${CA_CERTIFICATES_VERSION} \
# 	# git=${GIT_VERSION} \
# 	ca-certificates \
# 	git \
# 	; \
# 	apt-get clean; \
# 	rm -rf /var/lib/apt/lists/*; \
# 	git config --global http.sslverify false

# RUN env

# Copy in the hops executable
COPY ${HOPS_EXECUTABLE} /bin/hops

ENV PATH="/bin:${PATH}"
