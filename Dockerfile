FROM alpine:3.12
EXPOSE 8803
WORKDIR /app
COPY dist/app /app/server
RUN chmod +x /app/server
CMD ["./server"]
