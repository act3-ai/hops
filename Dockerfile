# hadolint ignore=DL3007
FROM chainguard/wolfi-base:latest

# By default, depends on the artifact from the "build linux" job in the pipeline
ARG HOPS_EXECUTABLE=ci-dist/hops/linux/amd64/bin/hops

# Copy in the hops executable
COPY ${HOPS_EXECUTABLE} /bin/hops

# Add the hops executable to the PATH
ENV PATH="/bin:${PATH}"
