DROP TABLE IF EXISTS vmstat.metrics;
CREATE TABLE vmstat.metrics (
	datetime datetime,
	running int unsigned NOT NULL,
	blocking int unsigned NOT NULL,
	swapped int unsigned NOT NULL,
	free int unsigned NOT NULL,
	buffer int unsigned NOT NULL,
	cache int unsigned NOT NULL,
	swap_in int unsigned NOT NULL,
	swap_out int unsigned NOT NULL,
	block_in int unsigned NOT NULL,
	block_out int unsigned NOT NULL,
	interapt int unsigned NOT NULL,
	context_switch int unsigned NOT NULL,
	cpu_user tinyint unsigned NOT NULL,
	cpu_system tinyint unsigned NOT NULL,
	cpu_idle tinyint unsigned NOT NULL,
	cpu_iowait tinyint unsigned NOT NULL,
	cpu_steal tinyint unsigned NOT NULL,
	PRIMARY KEY(datetime)
);
