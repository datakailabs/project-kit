# Example .project.toml Configurations

This directory contains example `.project.toml` files demonstrating different project configurations.

## Examples

### minimal.toml
The bare minimum required configuration for a PK project. Use this as a starting point for simple projects.

### web-app.toml
Configuration for a typical web application with a custom tmux layout featuring editor, server, logs, and git windows.

### data-project.toml
Data engineering project with cloud context switching (AWS, Databricks, Snowflake) and a tmux layout optimized for data work.

### full-featured.toml
Comprehensive example showing all available configuration options including client info, multiple cloud contexts, and complex tmux layouts.

## Usage

Copy an example to your project directory:

```bash
cp docs/examples/minimal.toml ~/projects/myproject/.project.toml
```

Then edit it to match your project needs:

```bash
pk edit myproject
```

## Configuration Reference

See `man pk` for detailed documentation on all configuration options.
