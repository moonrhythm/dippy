build:
	buildctl build \
		--frontend dockerfile.v0 \
		--local dockerfile=. \
		--local context=. \
		--output type=image,name=registry.moonrhythm.io/dippy,push=true
