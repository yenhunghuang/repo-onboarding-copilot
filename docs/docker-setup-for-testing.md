# Docker Setup for Security Integration Testing

This document provides guidance for setting up Docker to run the full security integration test suite for the Repo Onboarding Copilot project.

## Prerequisites

- Docker Desktop installed and running
- Go 1.21+ installed
- Access to pull container images

## Docker Configuration for Tests

The security integration tests require Docker to create isolated containers for testing the security sandbox functionality. The tests use distroless images for enhanced security.

### Required Docker Setup

1. **Ensure Docker is Running**:
   ```bash
   docker --version
   docker info
   ```

2. **Pull Required Test Images**:
   ```bash
   docker pull gcr.io/distroless/static-debian12:latest
   ```

3. **Verify Docker Daemon Access**:
   ```bash
   docker run --rm hello-world
   ```

### Security Profile Configuration

The tests use security profiles that may need to be available:

1. **Create Security Configuration Directory**:
   ```bash
   mkdir -p configs/security
   ```

2. **Default Seccomp Profile** (Optional):
   The tests attempt to use a default seccomp profile. If you encounter "seccomp profile not found" errors, you can:
   
   - **Option A**: Disable seccomp for testing (less secure):
     ```bash
     # Tests will handle this gracefully and skip container-dependent tests
     ```
   
   - **Option B**: Create a basic seccomp profile:
     ```bash
     # Create configs/security/seccomp-analysis.json with default Docker seccomp
     curl -o configs/security/seccomp-analysis.json \
          https://raw.githubusercontent.com/docker/docker-ce/master/components/engine/profiles/seccomp/default.json
     ```

### Running Security Tests

1. **Run All Security Tests**:
   ```bash
   go test ./test/integration/security/... -v
   ```

2. **Run Container Security Tests Specifically**:
   ```bash
   go test ./test/integration/security -run "TestContainer" -v
   ```

3. **Run with Docker Debug Info**:
   ```bash
   DOCKER_API_VERSION=1.40 go test ./test/integration/security/... -v
   ```

## Test Behavior Without Docker

The security integration tests are designed to gracefully handle Docker unavailability:

- **Container Creation Tests**: Skipped with warning if Docker is unavailable
- **Security Policy Tests**: Run validation without actual container creation
- **Audit Logging Tests**: Run independently of container infrastructure
- **Resource Monitoring Tests**: Test monitoring logic without containers

This ensures the core security functionality can be tested even in CI environments where Docker may not be available.

## Troubleshooting

### Common Issues

1. **"docker: opening seccomp profile (default) failed"**:
   - This is expected if no custom seccomp profile is configured
   - Tests will skip container creation and continue with other validation
   - To resolve: Follow seccomp profile setup above

2. **"Cannot connect to the Docker daemon"**:
   - Ensure Docker Desktop is running
   - Check Docker daemon is accessible: `docker ps`

3. **"Permission denied" accessing Docker socket**:
   - On Linux: Add user to docker group: `sudo usermod -aG docker $USER`
   - Restart terminal session after group change

4. **Container fails to start**:
   - Check if the distroless image is available: `docker images | grep distroless`
   - Verify Docker has sufficient resources allocated

### CI/CD Environment

For automated testing environments:

```bash
# Example GitHub Actions setup
- name: Set up Docker
  uses: docker/setup-docker@v2

- name: Pull test images
  run: docker pull gcr.io/distroless/static-debian12:latest

- name: Run security tests
  run: go test ./test/integration/security/... -v
```

## Test Coverage

With proper Docker setup, you'll get full test coverage including:

- ✅ Container security validation
- ✅ Runtime security policies
- ✅ Network isolation testing
- ✅ Resource limit enforcement
- ✅ Security scanning integration
- ✅ Container lifecycle management

Without Docker, you'll still get:

- ✅ Security policy validation
- ✅ Audit logging completeness
- ✅ Resource monitoring logic
- ✅ Configuration validation
- ✅ Defense-in-depth architecture validation

The test suite is designed to be robust and provide valuable security validation regardless of Docker availability.