# TAI (Terminal AI) Contributor Guide

This document provides a high-level overview of the TAI project, its structure, and development guidelines for contributors.

## Project Overview

TAI is a terminal-based AI assistant designed to provide a seamless interface for interacting with various Large Language Model (LLM) providers. It supports two primary modes of operation:

1.  **REPL (Interactive) Mode**: A multi-turn chat interface for conversational AI.
2.  **One-shot Mode**: A command-line mode for single prompts, designed for scripting and automation.

The goal is to create a powerful, extensible, and user-friendly tool for developers to leverage AI directly from their terminal.

## Tech Stack

-   **Language**: Go (Golang)
-   **CLI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
-   **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Folder Structure

The repository is organized into the following core directories:

-   `cmd/tai/`: The main application entry point.
-   `internal/`: Contains all the private, core application logic.
    -   `cli/`: CLI framework integration and command handling.
    -   `ui/`: Bubble Tea components and view logic.
    -   `llm/`: Interfaces and implementations for LLM providers (OpenAI, Gemini, etc.).
    -   `tools/`: Logic for agentic tools like file system operations and shell commands.
    -   `state/`: Redux-style state management (store, actions, reducers).
-   `pkg/`: Public libraries and shared models intended for external use.
    -   `models/`: Core data structures.
    -   `utils/`: Utility functions.
-   `build/`: Stores compiled binaries (created during the build process).

## Development

All common development tasks are managed through the `Makefile`.

### Common Commands

-   **Build the application**:
    ```sh
    make build
    ```

-   **Run the application**:
    ```sh
    make run
    ```

-   **Run tests**:
    ```sh
    make test
    ```

-   **Run linters and formatters**:
    ```sh
    make check
    ```

-   **Clean build artifacts**:
    ```sh
    make clean
    ```

For a full list of commands, run `make help`.
