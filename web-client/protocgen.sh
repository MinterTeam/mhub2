#!/usr/bin/env bash

# Path to this plugin
PROTOC_GEN_TS_PATH="./node_modules/.bin/protoc-gen-ts"

# Directory to write generated code to (.js and .d.ts files)
OUT_DIR="./gen/"

rm -rf ${OUT_DIR}
mkdir -p ${OUT_DIR}
protoc \
-I "../module/proto" \
-I "../module/third_party/proto" \
--grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:${OUT_DIR} \
--js_out=import_style=commonjs:${OUT_DIR} \
$(find ../module -maxdepth 99 -name '*.proto')

#--plugin=${PROTOC_GEN_TS_PATH} \
#--js_out="import_style=commonjs:${OUT_DIR}" \
#--ts_out="service=grpc-web:${OUT_DIR}" \

