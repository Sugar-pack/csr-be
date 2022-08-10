## How to run this code at existing hosting
### Notice: this instruction for macOS.

1. Install musl-cross-compiler:
```shell
brew install FiloSottile/musl-cross/musl-cross
 ```
2. Build the application:
```shell
make linux-build-mac-version
```
3. Access the server via ssh:
```shell
ssh -i ~/.ssh/kot ftp_160623@shared-21076501-160623.pleskbox.com
```
5. Make sure that all migrations file exist at server. For now all our migration files are at /go/db/migrations folder.
6. Copy binary to server.
7. Export all needed environment variables.
8. Run the application.