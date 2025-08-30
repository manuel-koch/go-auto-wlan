#!/usr/bin/env python3
import subprocess
import sys

if __name__ == "__main__":
    cert_name = sys.argv[1]
    certificates = subprocess.check_output(["security","find-certificate", "-c",cert_name,"-a","-p"],encoding="utf8")
    for cert in certificates.replace("-----END CERTIFICATE-----","-----END CERTIFICATE-----<SPLITTER>").split("<SPLITTER>"):
        cert = cert .strip()
        if cert:
            pem = subprocess.check_output(["openssl", "x509", "-noout", "-text", "-inform", "pem"],encoding="utf8",input=cert)
            print(pem)
