# 代理 Makefile：根目录执行，命令转发到 src/backend/Makefile
.PHONY: help env up down restart logs ps backend frontend test migrate

help env up down restart logs ps backend frontend test migrate:
	@$(MAKE) -C src/backend $@
