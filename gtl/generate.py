#!/usr/bin/env python3.6

"""
Simple template engine.

Usage:

    generate.py [options] foo.go.tpl

Example:

    generate.py --prefix= -DELEM=int32 --package=tests --output=unsafe.go ../unsafe.go.tpl

--prefix=ARG replaces all occurrences of "ZZ" with "ARG".

--Dfrom=to replaces all occurrences of "from" with "to". This flag can be set multiple times.

--output=path specifies the output file name.

"""

import re
import argparse
import sys

def main() -> None:
    "Main application entry point"
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '--package', default='funkymonkeypackage',
        help="Occurrences of 'PACKAGE' in the template are replaced with this string.")
    parser.add_argument(
        '--prefix', default='funkymonkey',
        help="Occurrences of 'ZZ'  in the template are replaced with this string")
    parser.add_argument(
        '-o', '--output', default='',
        help="Output destination. Defaults to standard output")
    parser.add_argument(
        '-D', '--define', default=[],
        type=str, action='append',
        help="str=replacement")
    parser.add_argument(
        'template', help="*.go.tpl file to process")

    opts = parser.parse_args()

    if opts.output == '':
        out = sys.stdout
    else:
        out = open(opts.output, 'w')

    defines = []
    for d in opts.define:
        m = re.match("^([^=]+)=(.*)", d)
        if m is None:
            raise Exception("Invalid -D option: " + d)
        defines.append((m[1], m[2]))

    print('// Code generated from \"', ' '.join(sys.argv), '\". DO NOT EDIT.', file=out)
    for line in open(opts.template, 'r').readlines():
        line = line.replace('ZZ', opts.prefix)
        line = line.replace('PACKAGE', opts.package)
        for def_from, def_to in defines:
            line = line.replace(def_from, def_to)
        print(line, end='', file=out)

main()
