#!/bin/bash
systemctl daemon-reload
systemctl enable peekl-server.service
systemctl start peekl-server.service
