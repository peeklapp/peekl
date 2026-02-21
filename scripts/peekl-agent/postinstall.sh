#!/bin/bash
systemctl daemon-reload
systemctl enable peekl-agent.service
systemctl start peekl-agent.service
