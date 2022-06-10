#!/bin/bash

# Warning: Either use command line build or IDE build but avoid mixing both during your workflow, 
# as it can slow down build times
# See https://issuetracker.google.com/issues/164145066

trap ctrl_c INT

# Avoid say command on ctrl-c
function ctrl_c() {
    exit 1
}

set -x # show executing lines

./gradlew assembleInternalDebug "$@"
    
retVal=$?

set +x

if [ $retVal -ne 0 ]; then
    say not assembled &
    exit $retVal
fi


output=$(adb install -r -t ./app/build/outputs/apk/internal/debug/app-internal-arm64-v8a-debug.apk && \
    adb shell am start -n com.Slack.internal.debug/slack.features.home.HomeActivity 2>&1)

retVal=$?

if [ $retVal -eq 0 ]; then
    say install complete &
else
    say "$output" &
    exit $retVal
fi

date
exit $retVal
