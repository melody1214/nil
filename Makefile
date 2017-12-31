OPENSSL		?= openssl

ROOTCA_DIR	:= pkg/security/test_certs
ROOTCA_KEY	:= $(ROOTCA_DIR)/rootCA.key
ROOTCA_PEM	:= $(ROOTCA_DIR)/rootCA.pem

CERTS_DIR	:= .certs
SERVER_KEY	:= $(CERTS_DIR)/test-server.key
SERVER_CSR	:= $(CERTS_DIR)/test-server.csr
SERVER_CRT	:= $(CERTS_DIR)/test-server.crt
T_ROOTCA_PEM	:= $(CERTS_DIR)/test-rootCA.pem

.PHONY: certs
certs:
	mkdir -p $(CERTS_DIR)
	cp $(ROOTCA_PEM) $(T_ROOTCA_PEM)
	${OPENSSL} genrsa -out $(SERVER_KEY) 2048
	${OPENSSL} req -new -key $(SERVER_KEY) -out $(SERVER_CSR)
	${OPENSSL} x509 -req -days 3650 -in $(SERVER_CSR) \
		-CA $(T_ROOTCA_PEM) -CAcreateserial \
		-CAkey $(ROOTCA_KEY) \
		-out $(SERVER_CRT)
	rm .srl

GO ?= go 

.DEFAULT_GOAL	:= all

.PHONY: all
all:
	$(GO) build

.PHONY: clean
clean:
	$(GO) clean
