#!/usr/bin/env python3

import bs4

heading_map = {
        'h1': '\\section',
        'h2': '\\subsection',
        'h3': '\\subsubsection',
        'h4': '\\paragraph',
        'h5': '\\subparagraph'
    }

def transform(elem):
    return ''.join(_transform_child(elem))

def _transform_child(tag):
    def esc(s, chars='&%$#_{}'):
        for c in chars:
            s = s.replace(c, f'\\{c}')
        return s

    in_itemize = False
    for elem in tag:
        if isinstance(elem, bs4.element.NavigableString):
            if str(elem).strip() and in_itemize:
                yield '\n'
                yield '\\end{itemize}'
                yield '\n'
                in_itemize = False
            yield str(elem)
            continue
        # child must be isinstance(bs4.element.Tag)

        if elem.name == 'ul':
            if not in_itemize:
                yield '\n'
                yield '\\begin{itemize}'
                yield '\n'
                in_itemize = True # keep track of this to remove redundand end tags. The HTML is bad here.
            for child in elem:
                if isinstance(child, bs4.element.NavigableString):
                    if str(child).strip():
                        yield f'% Unexpected text in <ul>: {child}'
                elif not child.name == 'li':
                    yield f'% Unexpected element in <ul>: {child}'
                else: # <li>
                    for foo  in child:
                        if isinstance(foo, bs4.element.NavigableString):
                            if str(foo).strip():
                                yield '\n'
                                yield '\item '
                                break
                            # skip empty
                        elif foo.name == 'ul':
                            break
                        else:
                            yield '\n'
                            yield '\item '
                            break
                    yield from _transform_child(child)
            continue
        else:
            if in_itemize:
                yield '\n'
                yield '\\end{itemize}'
                yield '\n'
                in_itemize = False

        if elem.name in heading_map:
            # $ $ is a hack that seems necessary since otherwise the \paragraphs above the listings at the end don't
            # show up in the output
            yield f'{heading_map[elem.name]}{{{esc(elem.text)}}}$ $'
        elif elem.name == 'p':
            if 'proverif' in elem.attrs.get('class', []):
                yield '\\begin{lstlisting}'
                yield '\n'
                yield esc(elem.text, chars='{}')
                yield '\n'
                yield '\\end{lstlisting}'
                yield '\n'
                pass
            else:
                yield from _transform_child(elem)
                yield '\n\n'
        elif elem.name == 'em':
            yield f'\\emph{{{esc(elem.text)}}}'
        elif elem.name == 'span':
            yield f'\\texttt{{{esc(elem.text)}}}'
        else:
            yield '\n'
            yield f'% Unhandled tag {elem.name} {elem.attrs}'
            yield elem.text

    if in_itemize:
        yield '\n\\end{itemize}\n'


if __name__ == '__main__':
    import os
    import os.path as path
    import re
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('indir', default='-')
    parser.add_argument('output', type=argparse.FileType('w'), default='-')
    args = parser.parse_args()

    with args.output as out:
        with open(path.join(args.indir, 'index.html'), 'r') as f:
            soup = bs4.BeautifulSoup(f.read(), 'lxml')

        out.write(f'\\title{{{soup.find("title").text}}}')
        out.write('\\date{\\today}')
        out.write('\\begin{document}')
        out.write('\\maketitle')

        for infile in sorted(os.listdir(args.indir)):
            if not re.match('[A-Z]\.html?', infile):
                continue
            with open(path.join(args.indir, infile), 'r') as f:
                soup = bs4.BeautifulSoup(f.read(), 'lxml')

            out.write(f'\\section{{{soup.find("title").text.partition("-")[2]}}}\n')
            out.write(transform(soup.find('div', class_='resultsExplanation')))

