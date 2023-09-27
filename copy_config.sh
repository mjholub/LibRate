#!/bin/sh

# Print the contents of example_config.yml, then ask the user if they want to
# copy it unchanged to config.yml or open $EDITOR to edit it.

cat example_config.yml
echo -n "Above is the example config file. Would you like to copy it to config.yml, edit it or skip? (c/e/s)"
read answer

switch $answer in:
  cp example_config.yml config.yml
  echo "Copied example_config.yml to config.yml"
  case e
  $EDITOR config.yml
  case s:
  echo "Skipping config.yml"
end
