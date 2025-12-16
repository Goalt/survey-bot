
ENV=dev TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=postgres DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres DB_MIGRATIONS_UP=true RELEASE_VERSION=UNKNOWN ./bin/cli survey-create /workspace/surveytests/1.json
ENV=dev TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=postgres DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres DB_MIGRATIONS_UP=true RELEASE_VERSION=UNKNOWN ./bin/cli survey-create /workspace/surveytests/2.json
ENV=dev TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=postgres DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres DB_MIGRATIONS_UP=true RELEASE_VERSION=UNKNOWN ./bin/cli survey-create /workspace/surveytests/3.json
ENV=dev TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=postgres DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres DB_MIGRATIONS_UP=true RELEASE_VERSION=UNKNOWN ./bin/cli survey-create /workspace/surveytests/4.json
ENV=dev TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=postgres DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres DB_MIGRATIONS_UP=true RELEASE_VERSION=UNKNOWN ./bin/cli survey-create /workspace/surveytests/5.json


ENV=prod TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres RELEASE_VERSION=UNKNOWN ./bin/cli survey-update aea4e17a-19ca-4126-9481-952e6d304ed8 /workspace/surveytests/5.json

ENV=prod TOKEN=5924083863:AAGqxdX3i25VwmIpzlEhS_VJsiG-MJw53DY DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PWD=postgres DB_NAME=postgres RELEASE_VERSION=UNKNOWN ./bin/cli survey-update b55c5c96-d731-497c-a8bb-28c8915f6072 /workspace/surveytests/6.json
