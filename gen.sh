#!/bin/bash
[ ! -e proto ] && mkdir proto
protoc -I . proto.proto --go_out=plugins=grpc:proto
