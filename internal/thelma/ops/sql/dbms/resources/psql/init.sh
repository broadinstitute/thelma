#!/bin/bash

#
# This script provisions the thelma-sql-ro and thelma-sql-rw users inside the
# Postgres database.
#

set -euo pipefail

CLOUDSQL_DBOWNER="cloudsqlsuperuser"

run_psql_no_echo() {
  psql --no-psqlrc --quiet --set=ON_ERROR_STOP=1 "$@"
}

run_psql() {
  run_psql_no_echo --echo-all "$@"
}

# return 0 if user exists, 1 otherwise
user_exists() {
  user="$1"
  run_psql_no_echo -t --csv -c "select rolname from pg_roles WHERE rolname = '${user}';" | grep -F "${user}"
}

# Print list of Terra databases in this instance to stdout, one on each line
list_dbs_and_owners() {
  run_psql_no_echo -t --csv -c \
    "select d.datname, pg_catalog.pg_get_userbyid(d.datdba) \
     from pg_catalog.pg_database d \
     where d.datistemplate=false and \
     d.datname not in ('postgres', 'cloudsqladmin');"
}

regex="^(un|re)?init$"
if [[ "$#" != 1 ]] || ! [[ "${1}" =~ $regex ]]; then
  cat <<EOF
Usage: $0 [init|uninit|reinit]

This script initializes a Postgres database for Thelma connections,
including creating read-only and read-write users. It has 3 modes:

init   - create users & grant permissions
uninit - revoke permissions & delete users (re-runnable)
reinit - uninit followed by init

EOF

  exit 1
fi

action="${1}"

if [[ "${action}" == "uninit" ]] || [[ "${action}" == "reinit" ]]; then
  ro_user_exists=false
  if user_exists "${INIT_RO_USER}"; then
    ro_user_exists=true
  fi

  rw_user_exists=false
  if user_exists "${INIT_RW_USER}"; then
    rw_user_exists=true
  fi

  # Loop through databases and revoke perms
  OLDIFS=$IFS
  IFS=','
  list_dbs_and_owners | while read db owner; do
    echo "> Updating permissions for database ${db} (owner ${owner})"
    # Grant readwrite account the same permissions as app user
    # ref: https://stackoverflow.com/a/19602050

    # So in CloudSQL, the database owner for all databases is
    # "cloudsqlsuperuser" but the application user (eg.
    # "workspacemanager-landingzone")
    # is the table owner.
    #
    # Because we are modifying table permissions, we need
    # to assume the table owner user's permissions.
    #
    # Conventionally in CloudSQL, the table owner role has
    # the same name as the database.
    #
    if [[ "${owner}" == "${CLOUDSQL_DBOWNER}" ]]; then
      echo "This looks like a cloudsql database; assuming db owner is ${db}"
      owner="${db}"
    fi

    if [[ "${ro_user_exists}" == "true" ]]; then
      # Revoke readonly account permissions
      echo
      echo "> Revoking ${INIT_RO_USER} permissions for ${db}"
      run_psql -d "${db}" -c "revoke all privileges on database \"${db}\" from \"${INIT_RO_USER}\";"
      run_psql -d "${db}" -c "revoke all privileges on schema public from \"${INIT_RO_USER}\";"

      if [[ "${PGUSER}" != "${owner}" ]]; then
        echo "> Temporarily granting ${owner} role to ${PGUSER}"
        run_psql -d "${db}" -c "grant \"${owner}\" to \"${PGUSER}\";"
      fi
      echo "> Revoking privileges on all tables in ${db} from ${INIT_RO_USER}"
      run_psql -d "${db}" -c "revoke all privileges on all tables in schema public from \"${INIT_RO_USER}\";"
      if [[ "${PGUSER}" != "${owner}" ]]; then
        echo "> Revoking ${owner} role from ${PGUSER}"
        run_psql -d "${db}" -c "revoke \"${owner}\" from \"${PGUSER}\";"
      fi

      run_psql -d "${db}" -c "alter default privileges in schema public revoke all on tables from \"${INIT_RO_USER}\""
    fi

    if [[ "${rw_user_exists}" == "true" ]]; then
      # Revoke readwrite account permissions
      echo
      echo "> Revoking ${INIT_RW_USER} permissions for ${db}"
      run_psql -d "${db}" -c "revoke \"${owner}\" from \"${INIT_RW_USER}\";"
    fi
  done
  IFS=$OLDIFS

  # Delete thelma users if enabled
  if [[ "${INIT_CREATE_USERS}" == 'true' ]]; then
    echo
    echo "> Deleting Thelma users"
    run_psql -c "drop role if exists \"${INIT_RO_USER}\";"
    run_psql -c "drop role if exists \"${INIT_RW_USER}\";"
  fi
fi

if [[ "${action}" == "init" ]] || [[ "${action}" == "reinit" ]]; then
  # Create thelma users if enabled
  if [[ "${INIT_CREATE_USERS}" == 'true' ]]; then
    echo
    echo "> Creating Thelma users"
    run_psql -c "create role \"${INIT_RO_USER}\" with login;"
    run_psql -c "create role \"${INIT_RW_USER}\" with login;"

    # Set password if specified
    if [[ "${INIT_RO_PASSWORD}" != '' ]]; then
      echo
      echo "> Setting password for ${INIT_RO_USER}..."
      run_psql_no_echo -c "alter role \"${INIT_RO_USER}\" with password '${INIT_RO_PASSWORD}'"
    fi
    if [[ "${INIT_RW_PASSWORD}" != '' ]]; then
      echo
      echo "> Setting password for ${INIT_RO_USER}..."
      run_psql_no_echo -c "alter role \"${INIT_RW_USER}\" with password '${INIT_RW_PASSWORD}'"
    fi
  fi

  # Loop through databases and set appropriate perms
  OLDIFS=$IFS
  IFS=','
  list_dbs_and_owners | while read db owner; do
    echo "> Updating permissions for database ${db} (owner ${owner})"
    if [[ "${owner}" == "${CLOUDSQL_DBOWNER}" ]]; then
      echo "This looks like a cloudsql database; assuming db owner is ${db}"
      owner="${db}"
    fi

    # Grant readonly account permissions
    # ref: https://stackoverflow.com/a/42044878
    echo
    echo "> Granting ${INIT_RO_USER} permissions for ${db}"
    run_psql -d "${db}" -c "grant connect on database \"${db}\" to \"${INIT_RO_USER}\";"
    run_psql -d "${db}" -c "grant usage on schema public to \"${INIT_RO_USER}\";"

    if [[ "${PGUSER}" != "${owner}" ]]; then
      echo "> Temporarily granting ${owner} role to ${PGUSER}"
      run_psql -d "${db}" -c "grant \"${owner}\" to \"${PGUSER}\";"
    fi
    echo "> Granting privileges on all tables in ${db} to ${INIT_RO_USER}"
    run_psql -d "${db}" -c "grant select on all tables in schema public to \"${INIT_RO_USER}\";"
    if [[ "${PGUSER}" != "${owner}" ]]; then
      echo "> Revoking ${owner} role from ${PGUSER}"
      run_psql -d "${db}" -c "revoke \"${owner}\" from \"${PGUSER}\";"
    fi

    run_psql -d "${db}" -c "alter default privileges in schema public grant select on tables to \"${INIT_RO_USER}\""

    # Grant readwrite account requisite permissions
    # ref: https://stackoverflow.com/a/19602050
    echo
    echo "> Granting ${INIT_RW_USER} permissions for ${db}"
    run_psql -d "${db}" -c "grant \"${owner}\" to \"${INIT_RW_USER}\";"
  done
fi
IFS=$OLDIFS