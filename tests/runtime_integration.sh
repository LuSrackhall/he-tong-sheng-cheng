#!/bin/bash
# Knowledge Runtime Integration Test
# Runs kr plan against all capabilities and core workflows.

set -e
echo "=== Knowledge Runtime Integration Test ==="
echo ""

# Build CLI
go build -o kr ./runtime/cmd/kr/ 2>/dev/null

# Test all capabilities
echo "--- Testing kr plan (all capabilities) ---"
for cap in login collect-rent create-contract issue-receipt backup-database create-user ensure-contract-active; do
  echo -n "  $cap ... "
  output=$(./kr plan "$cap" 2>&1)
  if echo "$output" | grep -q "Plan:"; then
    echo "PASS"
  else
    echo "FAIL"
    echo "$output"
    exit 1
  fi
done

# Test all workflows
echo ""
echo "--- Testing kr plan (workflows) ---"
for wf in sign-new-contract renew-contract; do
  echo -n "  $wf ... "
  output=$(./kr plan "$wf" 2>&1)
  if echo "$output" | grep -q "Steps:"; then
    echo "PASS"
  else
    echo "FAIL"
    echo "$output"
    exit 1
  fi
done

# Test kr run (dry run with --validate disabled)
echo ""
echo "--- Testing kr run (dry run, no backend) ---"
echo -n "  collect-rent ... "
output=$(./kr run collect-rent --validate=false 2>&1)
if echo "$output" | grep -q "Execution complete"; then
  echo "PASS"
else
  echo "FAIL"
  echo "$output"
  exit 1
fi

# Test kr explain
echo -n "  explain ... "
trace_id=$(echo "$output" | grep "Trace:" | awk '{print $NF}')
output=$(./kr explain "$trace_id" 2>&1)
if echo "$output" | grep -q "Capability:"; then
  echo "PASS"
else
  echo "FAIL"
  echo "$output"
  exit 1
fi

echo ""
echo "=== All integration tests PASSED ==="
rm -f kr
