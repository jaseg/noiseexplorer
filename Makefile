
PATTERN_DIRS := $(shell find html/patterns -mindepth 1 -maxdepth 1 -type d)
PDF_TARGETS := $(patsubst html/patterns/%,pdf/%.pdf,$(PATTERN_DIRS))


all: dirs pdfs

.PHONY: dirs
dirs:
	mkdir -p tex pdf tex_tmp

.PHONY: pdfs
pdfs: $(PDF_TARGETS)


pdf/%.pdf: tex/%.tex
	cp -f $^ tex_build/in.tex
	cd tex_build; pdflatex template.tex
	mv tex_build/template.pdf $@

.PRECIOUS: tex/%.tex
tex/%.tex: html/patterns/%
	python3 html_to_tex.py $^ $@

