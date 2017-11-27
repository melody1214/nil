PROTOC		?= protoc
PROTOCFLAGS	:= -I .:../../../: --go_out=plugins=grpc:.

PROTOS		:= SWIM MDS
SWIM_PROTO	:= pkg/swim/swimpb/swim.proto
MDS_PROTO	:= pkg/mds/mdspb/mds.proto

.PHONY: proto
proto: $(PROTOS)
$(PROTOS):
	${PROTOC} ${PROTOCFLAGS} $($@_PROTO)

OPENSSL		?= openssl

ROOTCA_DIR	:= pkg/security/test_certs
ROOTCA_KEY	:= $(ROOTCA_DIR)/rootCA.key
ROOTCA_PEM	:= $(ROOTCA_DIR)/rootCA.pem

CERTS_DIR	:= .certs
SERVER_KEY	:= $(CERTS_DIR)/test-server.key
SERVER_CRT	:= $(CERTS_DIR)/test-server.crt
T_ROOTCA_PEM	:= $(CERTS_DIR)/test-rootCA.pem

.PHONY: certs
certs:
	mkdir -p $(CERTS_DIR)
	cp $(ROOTCA_PEM) $(T_ROOTCA_PEM)
	${OPENSSL} genrsa -out $(SERVER_KEY) 2048
	${OPENSSL} req -x509 -new -nodes -key $(ROOTCA_KEY) -sha256 -days 3650 -out $(SERVER_CRT)
