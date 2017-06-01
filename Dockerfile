FROM scratch

COPY dist/linux_amd64/aliyundisk-provisioner /usr/bin/

ENTRYPOINT ["/usr/bin/aliyundisk-provisioner"]
