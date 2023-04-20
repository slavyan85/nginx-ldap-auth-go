# nginx-ldap-auth-go

## Service for handle Nginx internal auth request with LDAP

How to use:
- build cmd/ldap-auth.go and put binary to /usr/local/sbin/ldap-auth
- create /etc/systemd/system/ldap-auth.service

      [Unit]
      Description=Nginx LDAP auth backend 
      After=network.target
      Requires=network.target
      
      [Service]
      Type=simple
      ExecStart=/usr/local/sbin/ldap-auth --config /etc/ldap-auth/config.yml
      Restart=always
      
      [Install]
      WantedBy=multi-user.target
- create /etc/ldap-auth/config.yml

      ---
      bind: "127.0.0.1:8888"
      cookieName: "ldap-auth"
      debug: no
      url: "/ldap-auth"
      ldap:
        address: "ldap1.srv:636"
        base: "OU=users,DC=domain,DC=srv"
        bind:
          user: "SRV\\ldap-servic-user"
          password: "ldap-servic-user-password"
        filter:
          user: "(sAMAccountName=%s)"
          group: ""
        ssl:
          use: yes
          skipTls: no
          skipVerify: yes
          serverName: "ldap.srv"
- configure Nginx to use auth_request

      proxy_cache_path    cache/  keys_zone=auth_cache:10m;
      
      server {
        listen 80;
        autoindex                   on;
        autoindex_exact_size        off;
        autoindex_localtime         on;
        charset                     utf-8;
      
        location /ldap-auth {
          internal;
          proxy_pass      http://127.0.0.1:8888;
          proxy_pass_request_body off;
          proxy_set_header Content-Length "";
          proxy_cache auth_cache;
          proxy_cache_valid 200 30d;
          proxy_cache_key "$http_authorization$cookie_nginxauth";
          proxy_set_header X-Real-IP $remote_addr;
          proxy_set_header Cookie ldap-auth=$cookie_nginxauth;
        }
      
        location / {
          auth_request            /ldap-auth;
          alias                   /var/www/;
        }
      }