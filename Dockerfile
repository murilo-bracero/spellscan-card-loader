FROM alpine:3.16

WORKDIR /app

RUN chown 1001 /app \
    && chmod "g+rwX" /app \
    && chown 1001:root /app

COPY --chown=1001:root build/spellscan-card-loader spellscan-card-loader

USER 1001

CMD ["./spellscan-card-loader"]