FROM alpine:latest

ARG BUILD_DATE

# first, a bit about this container
LABEL build_info="mr-ingenious/docker-ntp build-date:- ${BUILD_DATE}"
LABEL maintainer="Mr. Ingenious <mr.ingenious@gmail.com>"
LABEL documentation="https://github.com/mr-ingenious/docker-ntp"

# default configuration
ENV NTP_DIRECTIVES="ratelimit\nrtcsync"

# install chrony
RUN apk add --no-cache chrony tzdata logrotate && \
    rm /etc/chrony/chrony.conf && \
    mkdir /run/chrony /opt/www /opt/www/res && \
    chown -R chrony:chrony /etc/chrony /run/chrony /var/lib/chrony /var/log && \
    chmod 1750 /etc/chrony /run/chrony /var/lib/chrony

# download bootstrap
ADD https://cdn.jsdelivr.net/npm/bootstrap@5.3.8/dist/css/bootstrap.min.css /opt/www/
ADD https://cdn.jsdelivr.net/npm/bootstrap@5.3.8/dist/js/bootstrap.bundle.min.js /opt/www/

RUN chmod 0444 /opt/www/bootstrap.min.css /opt/www/bootstrap.bundle.min.js

# script to configure/startup chrony (ntp)
COPY --chmod=0755 assets/startup.sh /opt/startup.sh

# webserver
COPY --chmod=0755 assets/web/webserver /bin/webserver
COPY --chmod=0444 assets/web/static/index.html \
                  assets/web/static/tracking.html \
                  assets/web/static/home.html  \
                  assets/web/static/source_stats.html \
                  assets/web/static/server_stats.html \
                  assets/web/static/system.html \
                  assets/web/static/clients.html \
                  assets/web/static/sources.html \
                  assets/web/static/styles.css /opt/www
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
