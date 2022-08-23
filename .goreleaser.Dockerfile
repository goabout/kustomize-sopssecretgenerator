# Kustomize will run KRM Function containers as user nobody. Copy /etc/password
# from an Alpine image to make this possible.

FROM alpine:3 as alpine

FROM scratch
USER nobody
COPY --from=alpine /etc/passwd /etc/passwd
ENTRYPOINT ["/SopsSecretGenerator"]
COPY SopsSecretGenerator /
