FROM ubuntu:latest

COPY social-book-operator .

ENTRYPOINT [ "./social-book-operator" ]