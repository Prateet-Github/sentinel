package controlplane

const schema = `
CREATE TABLE IF NOT EXISTS services (
    name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS backends (
    name TEXT NOT NULL,
    service_name TEXT NOT NULL,
    url TEXT NOT NULL,
    health_check_path TEXT NOT NULL,

    PRIMARY KEY (service_name, name),

    FOREIGN KEY (service_name)
        REFERENCES services(name)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS routes (
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    service_name TEXT NOT NULL,

    PRIMARY KEY (method, path),

    FOREIGN KEY (service_name)
        REFERENCES services(name)
        ON DELETE CASCADE
);
`
