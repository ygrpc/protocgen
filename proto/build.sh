#!/bin/bash

protoc --ygrpc-protocsave_out=. --ygrpc-protocsave_opt=filename=aa.out --ygrpc-protocsave_opt=--protocgen-port=8000 t.proto