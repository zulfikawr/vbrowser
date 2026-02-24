# Contributing to vbrowser

Thank you for your interest in contributing to vbrowser! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**Good bug reports include:**
- Clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Environment details (OS, Go version, GStreamer version)
- Relevant logs (use `--foreground --log-level debug`)

### Suggesting Features

Feature requests are welcome! Please:
- Check if the feature already exists or is planned
- Clearly describe the use case and benefits
- Consider implementation complexity

### Pull Requests

1. **Fork the repository** and create a feature branch
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow existing code style
   - Add tests if applicable
   - Update documentation

3. **Test your changes**
   ```bash
   make test
   make lint
   ```

4. **Commit with clear messages**
   ```bash
   git commit -m "feat: add hardware encoding support"
   ```
   
   Use conventional commits:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `perf:` - Performance improvements
   - `refactor:` - Code refactoring
   - `test:` - Test additions/changes
   - `chore:` - Maintenance tasks

5. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

## Development Setup

### Prerequisites
```bash
# Ubuntu/Debian
sudo apt-get install -y \
    xvfb xdotool pulseaudio \
    libgstreamer1.0-dev \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-plugins-ugly \
    gstreamer1.0-pulseaudio

# Go 1.21+
go version
```

### Building
```bash
make build
```

### Running Tests
```bash
make test
```

### Linting
```bash
make lint
```

## Project Structure

```
vbrowser/
├── cmd/vbrowser/          # Main entry point
├── internal/
│   ├── browser/           # Chromium management
│   ├── config/            # Configuration
│   ├── platform/          # Platform-specific code
│   ├── process/           # PID management
│   └── stream/            # WebRTC streaming
├── pkg/
│   ├── gst/               # GStreamer bindings
│   ├── server/            # HTTP/WebSocket server
│   └── xorg/              # X11 input handling
└── configs/               # Example configs
```

## Coding Guidelines

### Go Code
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions small and focused
- Add comments for exported functions
- Handle errors explicitly

### JavaScript Code
- Use modern ES6+ syntax
- Keep functions pure when possible
- Add comments for complex logic
- Use meaningful variable names

### Commit Messages
- Use present tense ("add feature" not "added feature")
- Keep first line under 72 characters
- Reference issues when applicable (#123)

## Performance Considerations

When contributing performance improvements:
- Measure before and after with benchmarks
- Document the improvement in the PR
- Consider trade-offs (latency vs quality, CPU vs memory)
- Test on different hardware configurations

## Testing

- Add unit tests for new functionality
- Test on Linux (primary target)
- Verify with different GStreamer versions
- Check for memory leaks with long-running sessions

## Documentation

- Update README.md for user-facing changes
- Add inline code comments for complex logic
- Update CHANGELOG.md following semver
- Include examples for new features

## Questions?

- Open a GitHub issue for questions
- Check existing issues and discussions
- Review the README and documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to vbrowser! 🚀
