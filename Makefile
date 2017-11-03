PROTOC		?= protoc
PROTOCFLAGS	:= -I .:../../../: --go_out=plugins=grpc:.

PROTOS		:= SWIM MDS
SWIM_PROTO	:= pkg/swim/swim.proto
MDS_PROTO	:= pkg/mds/mdspb/mds.proto

.PHONY: proto
proto: $(PROTOS)
$(PROTOS):
	${PROTOC} ${PROTOCFLAGS} $($@_PROTO)