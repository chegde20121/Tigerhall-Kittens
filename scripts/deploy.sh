#!/bin/bash

# Default values
RUN_TESTS=true
APP_NAME="tigerhall_kitten_api"

# Print script usage
usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -notest          Skip running tests"
    echo "  -build           Skip tests and only build the project"
    echo "  -run             Skip tests and only run the server"
    echo "  -shutdown        Gracefully shutdown the running server"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        -notest)
            RUN_TESTS=false
            ;;
        -build)
            RUN_TESTS=false
            RUN_SERVER=false
            ;;
        -run)
            RUN_TESTS=false
            ;;
        -shutdown)
            SHUTDOWN_SERVER=true
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
    shift
done



# Run tests if the flag is set
if [ "$RUN_TESTS" = true ]; then
    echo "Running tests..."
    go test ./tests/unit_tests/... -coverprofile cover.out
fi

# Build the project
if [ "$RUN_SERVER" != false ]; then
    echo "Building the project..."
    go build -o $APP_NAME

    # Start or gracefully shutdown the server based on the flag
    if [ "$SHUTDOWN_SERVER" = true ]; then
        echo "Initiating graceful shutdown..."
        pkill -SIGINT -f ./$APP_NAME
    else
        echo "Starting the server..."
        ./$APP_NAME
    fi
fi
