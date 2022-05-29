.PHONY: *

BINARY_NAME=configurefs
TMP_DIR=/tmp/configurefs

configurefs:
	go build -o ${BINARY_NAME} cmd/configurefs/main.go

setup_test: configurefs
	mkdir -p ${TMP_DIR}/template_dir ${TMP_DIR}/mount_dir
	echo "{{.Value}}" > ${TMP_DIR}/template_dir/file
	echo "Value: value" > ${TMP_DIR}/variable_file
	echo "Run './configurefs -debug -mount-dir ${TMP_DIR}/mount_dir -template-dir ${TMP_DIR}/template_dir -variable-file ${TMP_DIR}/variable_file 2>&1 | jq' to test"

clean:
	go clean
	rm -rf ${BINARY_NAME} ${TMP_DIR}/template_dir ${TMP_DIR}/mount_dir ${TMP_DIR}/variable_file
