build:
	buildctl build \
		--frontend dockerfile.v0 \
		--local dockerfile=. \
		--local context=. \
		--output type=image,name=gcr.io/moonrhythm-containers/dippy,push=true
