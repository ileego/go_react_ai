# 根目录聚合 Makefile
# 所有命令统一转发到 src/Makefile，保持前后端命令入口一致

.PHONY: help

help:
	@$(MAKE) -C src help

%:
	@$(MAKE) -C src $@
