SERVICES = echo
PROTOC_INCLUDES= -Iproto -Iproto/3rdparty/api-common-protos -Iproto/3rdparty/protobuf -Iproto/3rdparty/protoc-gen-validate

gen:
	@$(foreach SERVICE,$(SERVICES), protoc $(PROTOC_INCLUDES)\
		--go_out=Mgoogle/protobuf/wrappers.proto,plugins=grpc,paths=source_relative:examples/api \
		--grpc-gateway_out=logtostderr=true,allow_patch_feature=false,paths=source_relative:examples/api \
		"proto/$(SERVICE)".proto;)
