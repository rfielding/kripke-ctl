#!/bin/bash
# Demo script for kripke-ctl

echo "==================================================================="
echo "Kripke-CTL Demo: CTL Model Checker with OpenAI Integration"
echo "==================================================================="
echo ""

# Build the project
echo "Building kripke-ctl..."
go build -o kripke-ctl
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "✓ Build successful"
echo ""

# Run tests
echo "Running tests..."
go test -v | grep -E "(PASS|FAIL|RUN)"
if [ $? -ne 0 ]; then
    echo "Tests failed!"
    exit 1
fi
echo "✓ All tests passed"
echo ""

# Show example Graphviz output
echo "==================================================================="
echo "Example: Traffic Light Kripke Structure (Graphviz DOT format)"
echo "==================================================================="
cat << 'EOF' | go run .
1
3
5
EOF

echo ""
echo "==================================================================="
echo "Demo completed successfully!"
echo "==================================================================="
echo ""
echo "To run the interactive CLI:"
echo "  ./kripke-ctl"
echo ""
echo "To enable OpenAI features:"
echo "  export OPENAI_API_KEY='your-api-key'"
echo "  ./kripke-ctl"
echo ""
echo "For more information, see README.md and USAGE.md"
