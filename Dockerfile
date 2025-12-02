FROM alpine:latest

ARG BUILD_DATE

# first, a bit about this container
LABEL build_info="mr-ingenious/docker-ntp build-date:- ${BUILD_DATE}"
LABEL maintainer="Mr. Ingenious <mr.ingenious@gmail.com>"
LABEL documentation="https://github.com/mr-ingenious/docker-ntp"

# default configuration
ENV NTP_DIRECTIVES="ratelimit\nrtcsync"

# install chrony
RUN apk add --no-cache chrony tzdata && \
    rm /etc/chrony/chrony.conf && \
    chmod 1750 /etc/chrony && \
    mkdir /run/chrony && \
    mkdir /opt/www && \
    mkdir /opt/www/res && \
    chown -R chrony:chrony /etc/chrony /run/chrony /var/lib/chrony && \
    chmod 1750 /etc/chrony /run/chrony /var/lib/chrony


# script to configure/startup chrony (ntp)
COPY --chmod=0755 assets/startup.sh /opt/startup.sh

# webserver
COPY --chmod=0755 assets/web/webserver /bin/webserver
COPY --chmod=0444 assets/web/static/index.html /opt/www/index.html
COPY --chmod=0444 assets/web/static/tracking.html /opt/www/tracking.html
COPY --chmod=0444 assets/web/static/home.html /opt/www/home.html
COPY --chmod=0444 assets/web/static/source_stats.html /opt/www/source_stats.html
COPY --chmod=0444 assets/web/static/server_stats.html /opt/www/server_stats.html
COPY --chmod=0444 assets/web/static/settings.html /opt/www/settings.html
COPY --chmod=0444 assets/web/static/clients.html /opt/www/clients.html
COPY --chmod=0444 assets/web/static/sources.html /opt/www/sources.html
COPY --chmod=0444 assets/web/static/styles.css /opt/www/styles.css
COPY --chmod=0444 assets/web/static/res/*.svg /opt/www/res

# ntp port
EXPOSE 123/udp

# web UI port
EXPOSE 80/tcp

VOLUME /etc/chrony /run/chrony /var/lib/chrony

# let docker know how to test container health
HEALTHCHECK CMD chronyc -n tracking || exit 1

# start chronyd in the foreground

USER chrony:chrony

ENTRYPOINT [ "/bin/sh", "/opt/startup.sh" ]
