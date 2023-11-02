#!/usr/bin/python3
#
# https://github.com/edoardottt/favirecon
#
# This script collects favicon hashes from the ProjectDiscovery/nuclei-templates GitHub repository.
#

import os
import yaml


# ----- retreive all nuclei templates -----
nuclei_templates_path = os.getenv('HOME') + "/nuclei-templates"

templates = [os.path.join(dp, f) for dp, dn, filenames in os.walk(nuclei_templates_path) for f in filenames if os.path.splitext(f)[1] == '.yaml']


# ----- scan all nuclei templates -----
for template in templates:
    with open(template, "r") as f:
        content = yaml.safe_load(f)
        if "metadata" in content["info"]:
            if "shodan-query" in content["info"]["metadata"]:
                if "favicon.hash" in content["info"]["metadata"]["shodan-query"]:
                    shodan_query = content["info"]["metadata"]["shodan-query"]
                    parts = shodan_query.split(":")
                    for part in parts:
                        if part.isnumeric() or part[0] == "-":
                            if "product" in content["info"]["metadata"]:
                                print(part + " " + content["info"]["metadata"]["product"])
                            else:
                                print("https://www.shodan.io/search?query=http.favicon.hash%3A" + part)