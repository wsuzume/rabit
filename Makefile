build_app:
	cd rabit-client && npm run build
	cd rabit-server && make build
	rm -rf bin
	mkdir -p bin
	cp -r rabit-client/dist bin
	cp rabit-server/server bin

run_app:
	cd bin && ./server
