#!/usr/bin/env python2

import errno
import os

import yaml
import jinja2
import fuse


fuse.fuse_python_api = (0, 2)


class MyStat(fuse.Stat):
    def __init__(self, st):
        self.st_mode = st.st_mode
        self.st_ino = st.st_mode
        self.st_dev = st.st_dev
        self.st_nlink = st.st_nlink
        self.st_uid = st.st_uid
        self.st_gid = st.st_gid
        self.st_size = st.st_size
        self.st_atime = st.st_atime
        self.st_mtime = st.st_mtime
        self.st_ctime = st.st_ctime


class ConfigureFS(fuse.Fuse):

    def __parse__(self):
        with open(self.vars_f) as f:
            content = f.read()
        context = yaml.load(content)
        return self.template.render(context)

    def __init__(self, src_f, vars_f, *args, **kw):

        fuse.Fuse.__init__(self, *args, **kw)

        self.src_f = src_f
        with open(src_f) as f:
            content = f.read()
        self.template = jinja2.Template(content)

        self.vars_f = vars_f

    def getattr(self, path):
        if path != "/":
            return - errno.ENOENT

        st = MyStat(os.stat(self.src_f))
        st.st_size = len(self.__parse__().encode('utf-8')) + 1000

        return st

    def open(self, path, flags):
        if path != "/":
            return -errno.ENOENT
        # PERMISSION CHECK IS MISSING

    def read(self, path, size, offset):
        t = self.__parse__()
        slen = len(t)
        if offset < slen:
            if offset + size > slen:
                size = slen - offset
            buf = t[offset:offset+size]
        else:
            buf = ''
        return buf.encode("utf-8")


def main():
        server = ConfigureFS("template", "vars")
        server.parse(errex=1)
        server.main()


if __name__ == '__main__':
    main()
