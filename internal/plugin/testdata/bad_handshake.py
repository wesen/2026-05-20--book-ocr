#!/usr/bin/env python3
"""Adversarial test plugin: contaminates stdout before any handshake."""
print("this is not a protocol frame")
import sys
sys.stdout.flush()
sys.stdin.read()
