-- =============================================================================
-- Rerunnable setup: drops and recreates DB_VIEWER_APP_DATA with seed presets
-- Compatible with Oracle 11g+ / XE / SQLcl / SQL*Plus
-- =============================================================================
SET DEFINE OFF;

-- ===================== DROP IF EXISTS =========================================
BEGIN EXECUTE IMMEDIATE 'DROP TABLE DB_VIEWER_APP_DATA CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN IF SQLCODE != -942 THEN RAISE; END IF; END;
/

-- ===================== TABLE: DB_VIEWER_APP_DATA ==============================
CREATE TABLE DB_VIEWER_APP_DATA (
    ID            NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    FEATURE       VARCHAR2(50)   NOT NULL,
    SCOPE_CLIENT  VARCHAR2(100),
    SCOPE_TABLE   VARCHAR2(100),
    ITEM_KEY      VARCHAR2(200),
    DATA          CLOB,
    CREATED_AT    TIMESTAMP DEFAULT SYSTIMESTAMP,
    UPDATED_AT    TIMESTAMP DEFAULT SYSTIMESTAMP
);

CREATE INDEX IDX_APPDATA_FEATURE ON DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE);
CREATE UNIQUE INDEX IDX_APPDATA_UNIQUE ON DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY);

-- =============================================================================
-- SEED DATA — Preset Filters & Queries from config.yml
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: CUSTOMERS
-- ─────────────────────────────────────────────────────────────────────────────

-- Preset Filters
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'CUSTOMERS', 'Personal Info',
    '{"name":"Personal Info","details":"Name, contact and location details","columns":["CUSTOMER_ID","FIRST_NAME","LAST_NAME","<SPACE>","<COMMENTARY> Contact","EMAIL","PHONE","<SPACE>","<COMMENTARY> Location","CITY","COUNTRY","<THE REST>"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'CUSTOMERS', 'Membership & Credit',
    '{"name":"Membership & Credit","details":"Membership tier and credit limit overview","columns":["CUSTOMER_ID","FIRST_NAME","LAST_NAME","<SPACE>","<COMMENTARY> Membership Details","MEMBERSHIP","CREDIT_LIMIT","REGISTERED_AT","<THE REST>"]}'
);

-- Preset Queries
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'CUSTOMERS', 'Search by Name',
    '{"name":"Search by Name","query":"SELECT * FROM {THIS_TABLE} WHERE UPPER(FIRST_NAME) LIKE ''%'' || UPPER(:NAME) || ''%'' OR UPPER(LAST_NAME) LIKE ''%'' || UPPER(:NAME) || ''%''","arguments":[{"name":"NAME","type":"string","description":"Enter first or last name to search"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'CUSTOMERS', 'By Country',
    '{"name":"By Country","query":"SELECT * FROM {THIS_TABLE} WHERE COUNTRY = :COUNTRY ORDER BY LAST_NAME","arguments":[{"name":"COUNTRY","type":"string","description":"Country code (e.g. USA, UK, Japan)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'CUSTOMERS', 'High Credit Customers',
    '{"name":"High Credit Customers","query":"SELECT * FROM {THIS_TABLE} WHERE CREDIT_LIMIT >= :MIN_CREDIT ORDER BY CREDIT_LIMIT DESC","arguments":[{"name":"MIN_CREDIT","type":"number","description":"Minimum credit limit"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'CUSTOMERS', 'By Membership Tier',
    '{"name":"By Membership Tier","query":"SELECT * FROM {THIS_TABLE} WHERE MEMBERSHIP = :TIER","arguments":[{"name":"TIER","type":"string","description":"STANDARD, SILVER, GOLD, or PLATINUM"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: PRODUCTS
-- ─────────────────────────────────────────────────────────────────────────────

-- Preset Filters
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'PRODUCTS', 'Catalog View',
    '{"name":"Catalog View","details":"Product name, SKU, price and availability","columns":["PRODUCT_ID","PRODUCT_NAME","SKU","<SPACE>","<COMMENTARY> Pricing & Stock","PRICE","STOCK_QTY","IS_AVAILABLE","<THE REST>"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'PRODUCTS', 'Logistics',
    '{"name":"Logistics","details":"Weight and category for shipping","columns":["PRODUCT_ID","PRODUCT_NAME","CATEGORY_ID","WEIGHT_KG","STOCK_QTY","<THE REST>"]}'
);

-- Preset Queries
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'PRODUCTS', 'Search by Product Name',
    '{"name":"Search by Product Name","query":"SELECT * FROM {THIS_TABLE} WHERE UPPER(PRODUCT_NAME) LIKE ''%'' || UPPER(:KEYWORD) || ''%''","arguments":[{"name":"KEYWORD","type":"string","description":"Keyword to search in product name"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'PRODUCTS', 'By Category',
    '{"name":"By Category","query":"SELECT * FROM {THIS_TABLE} WHERE CATEGORY_ID = :CAT_ID ORDER BY PRODUCT_NAME","arguments":[{"name":"CAT_ID","type":"number","description":"Category ID (1=Electronics, 2=Books, 3=Clothing, 4=Home, 5=Sports, 6=Discontinued)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'PRODUCTS', 'Price Range',
    '{"name":"Price Range","query":"SELECT * FROM {THIS_TABLE} WHERE PRICE BETWEEN :MIN_PRICE AND :MAX_PRICE ORDER BY PRICE","arguments":[{"name":"MIN_PRICE","type":"number","description":"Minimum price"},{"name":"MAX_PRICE","type":"number","description":"Maximum price"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'PRODUCTS', 'Low Stock Alert',
    '{"name":"Low Stock Alert","query":"SELECT * FROM {THIS_TABLE} WHERE STOCK_QTY <= :THRESHOLD AND IS_AVAILABLE = ''Y'' ORDER BY STOCK_QTY","arguments":[{"name":"THRESHOLD","type":"number","description":"Stock quantity threshold"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: ORDERS
-- ─────────────────────────────────────────────────────────────────────────────

-- Preset Filters
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'ORDERS', 'Order Summary',
    '{"name":"Order Summary","details":"Key order fields at a glance","columns":["ORDER_ID","CUSTOMER_ID","PRODUCT_ID","<SPACE>","<COMMENTARY> Financials","QUANTITY","UNIT_PRICE","TOTAL_AMOUNT","<SPACE>","<COMMENTARY> Status","STATUS","ORDER_DATE","<THE REST>"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'ORDERS', 'Shipping',
    '{"name":"Shipping","details":"Shipping and delivery tracking","columns":["ORDER_ID","STATUS","ORDER_DATE","SHIPPED_DATE","DELIVERY_NOTE","<THE REST>"]}'
);

-- Preset Queries
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'ORDERS', 'By Status',
    '{"name":"By Status","query":"SELECT * FROM {THIS_TABLE} WHERE STATUS = :STATUS ORDER BY ORDER_DATE DESC","arguments":[{"name":"STATUS","type":"string","description":"PENDING, CONFIRMED, SHIPPED, DELIVERED, or CANCELLED"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'ORDERS', 'By Customer',
    '{"name":"By Customer","query":"SELECT * FROM {THIS_TABLE} WHERE CUSTOMER_ID = :CUST_ID ORDER BY ORDER_DATE DESC","arguments":[{"name":"CUST_ID","type":"number","description":"Customer ID"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'ORDERS', 'Date Range',
    '{"name":"Date Range","query":"SELECT * FROM {THIS_TABLE} WHERE ORDER_DATE BETWEEN TO_DATE(:FROM_DATE, ''YYYY-MM-DD'') AND TO_DATE(:TO_DATE, ''YYYY-MM-DD'') ORDER BY ORDER_DATE","arguments":[{"name":"FROM_DATE","type":"string","description":"Start date (YYYY-MM-DD)"},{"name":"TO_DATE","type":"string","description":"End date (YYYY-MM-DD)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'ORDERS', 'Large Orders',
    '{"name":"Large Orders","query":"SELECT * FROM {THIS_TABLE} WHERE QUANTITY * UNIT_PRICE >= :MIN_TOTAL ORDER BY QUANTITY * UNIT_PRICE DESC","arguments":[{"name":"MIN_TOTAL","type":"number","description":"Minimum order total amount"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: EMPLOYEES
-- ─────────────────────────────────────────────────────────────────────────────

-- Preset Filters
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'EMPLOYEES', 'Directory',
    '{"name":"Directory","details":"Employee contact directory","columns":["EMPLOYEE_ID","FIRST_NAME","LAST_NAME","<SPACE>","<COMMENTARY> Contact","EMAIL","PHONE","<SPACE>","<COMMENTARY> Role","DEPARTMENT","JOB_TITLE","<THE REST>"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'EMPLOYEES', 'Compensation',
    '{"name":"Compensation","details":"Salary and hire date for review","columns":["EMPLOYEE_ID","FIRST_NAME","LAST_NAME","SALARY","HIRE_DATE","IS_ACTIVE","<THE REST>"]}'
);

-- Preset Queries
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'EMPLOYEES', 'By Department',
    '{"name":"By Department","query":"SELECT * FROM {THIS_TABLE} WHERE DEPARTMENT = :DEPT ORDER BY LAST_NAME","arguments":[{"name":"DEPT","type":"string","description":"Department name (Engineering, Sales, HR, Finance)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'EMPLOYEES', 'Salary Above',
    '{"name":"Salary Above","query":"SELECT * FROM {THIS_TABLE} WHERE SALARY >= :MIN_SAL ORDER BY SALARY DESC","arguments":[{"name":"MIN_SAL","type":"number","description":"Minimum salary"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'EMPLOYEES', 'Direct Reports',
    '{"name":"Direct Reports","query":"SELECT * FROM {THIS_TABLE} WHERE MANAGER_ID = :MGR_ID ORDER BY LAST_NAME","arguments":[{"name":"MGR_ID","type":"number","description":"Manager''s Employee ID"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'EMPLOYEES', 'Active Only',
    '{"name":"Active Only","query":"SELECT * FROM {THIS_TABLE} WHERE IS_ACTIVE = :ACTIVE","arguments":[{"name":"ACTIVE","type":"string","description":"Y for active, N for inactive"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: CATEGORIES
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'CATEGORIES', 'Active Categories',
    '{"name":"Active Categories","query":"SELECT * FROM {THIS_TABLE} WHERE IS_ACTIVE = :ACTIVE ORDER BY CATEGORY_NAME","arguments":[{"name":"ACTIVE","type":"string","description":"Y for active, N for inactive"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: AUDIT_LOG
-- ─────────────────────────────────────────────────────────────────────────────

-- Preset Filters
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'AUDIT_LOG', 'Change Summary',
    '{"name":"Change Summary","details":"Quick view of what changed","columns":["LOG_ID","TABLE_NAME","OPERATION","RECORD_ID","CHANGED_BY","CHANGED_AT","<THE REST>"]}'
);

-- Preset Queries
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'AUDIT_LOG', 'By Table',
    '{"name":"By Table","query":"SELECT * FROM {THIS_TABLE} WHERE TABLE_NAME = :TBL ORDER BY CHANGED_AT DESC","arguments":[{"name":"TBL","type":"string","description":"Table name (CUSTOMERS, PRODUCTS, ORDERS, EMPLOYEES)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'AUDIT_LOG', 'By Operation',
    '{"name":"By Operation","query":"SELECT * FROM {THIS_TABLE} WHERE OPERATION = :OP ORDER BY CHANGED_AT DESC","arguments":[{"name":"OP","type":"string","description":"INSERT, UPDATE, or DELETE"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'AUDIT_LOG', 'By User',
    '{"name":"By User","query":"SELECT * FROM {THIS_TABLE} WHERE CHANGED_BY = :USER_NAME ORDER BY CHANGED_AT DESC","arguments":[{"name":"USER_NAME","type":"string","description":"Username who made the change (e.g. admin, SYSTEM)"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: local_test / Table: DOCUMENTS
-- ─────────────────────────────────────────────────────────────────────────────

-- Preset Filters
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'local_test', 'DOCUMENTS', 'File Overview',
    '{"name":"File Overview","details":"Document name, type, size and uploader","columns":["DOCUMENT_ID","DOC_NAME","DOC_TYPE","FILE_SIZE","UPLOADED_BY","UPLOADED_AT","DESCRIPTION"]}'
);

-- Preset Queries
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'DOCUMENTS', 'By Type',
    '{"name":"By Type","query":"SELECT * FROM {THIS_TABLE} WHERE DOC_TYPE = :DOC_TYPE ORDER BY UPLOADED_AT DESC","arguments":[{"name":"DOC_TYPE","type":"string","description":"MIME type (e.g. text/csv, image/png)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'DOCUMENTS', 'By Uploader',
    '{"name":"By Uploader","query":"SELECT * FROM {THIS_TABLE} WHERE UPLOADED_BY = :UPLOADER ORDER BY UPLOADED_AT DESC","arguments":[{"name":"UPLOADER","type":"string","description":"Username who uploaded the file"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'local_test', 'DOCUMENTS', 'Large Files',
    '{"name":"Large Files","query":"SELECT * FROM {THIS_TABLE} WHERE FILE_SIZE >= :MIN_SIZE ORDER BY FILE_SIZE DESC","arguments":[{"name":"MIN_SIZE","type":"number","description":"Minimum file size in bytes"}]}'
);

-- =============================================================================
-- Client: sales_team
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: sales_team / Table: CUSTOMERS
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'sales_team', 'CUSTOMERS', 'Sales Contact',
    '{"name":"Sales Contact","details":"Customer contact info for sales outreach","columns":["CUSTOMER_ID","FIRST_NAME","LAST_NAME","EMAIL","PHONE","<SPACE>","<COMMENTARY> Membership","MEMBERSHIP","CREDIT_LIMIT"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'sales_team', 'CUSTOMERS', 'Gold & Platinum',
    '{"name":"Gold & Platinum","query":"SELECT * FROM {THIS_TABLE} WHERE MEMBERSHIP IN (''GOLD'', ''PLATINUM'') ORDER BY CREDIT_LIMIT DESC","arguments":[]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'sales_team', 'CUSTOMERS', 'By Country',
    '{"name":"By Country","query":"SELECT * FROM {THIS_TABLE} WHERE COUNTRY = :COUNTRY ORDER BY LAST_NAME","arguments":[{"name":"COUNTRY","type":"string","description":"Country code (e.g. USA, UK, Japan)"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: sales_team / Table: ORDERS
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'sales_team', 'ORDERS', 'Sales Pipeline',
    '{"name":"Sales Pipeline","details":"Order status and financials","columns":["ORDER_ID","CUSTOMER_ID","STATUS","<SPACE>","<COMMENTARY> Revenue","QUANTITY","UNIT_PRICE","TOTAL_AMOUNT","ORDER_DATE"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'sales_team', 'ORDERS', 'Pending Orders',
    '{"name":"Pending Orders","query":"SELECT * FROM {THIS_TABLE} WHERE STATUS IN (''PENDING'', ''CONFIRMED'') ORDER BY ORDER_DATE DESC","arguments":[]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'sales_team', 'ORDERS', 'Revenue Above',
    '{"name":"Revenue Above","query":"SELECT * FROM {THIS_TABLE} WHERE QUANTITY * UNIT_PRICE >= :MIN_TOTAL ORDER BY QUANTITY * UNIT_PRICE DESC","arguments":[{"name":"MIN_TOTAL","type":"number","description":"Minimum order total"}]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: sales_team / Table: PRODUCTS
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'sales_team', 'PRODUCTS', 'Price List',
    '{"name":"Price List","details":"Product pricing for quotes","columns":["PRODUCT_ID","PRODUCT_NAME","SKU","PRICE","IS_AVAILABLE"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'sales_team', 'PRODUCTS', 'Available Products',
    '{"name":"Available Products","query":"SELECT * FROM {THIS_TABLE} WHERE IS_AVAILABLE = ''Y'' ORDER BY PRODUCT_NAME","arguments":[]}'
);

-- =============================================================================
-- Client: hr_team
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: hr_team / Table: EMPLOYEES
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'hr_team', 'EMPLOYEES', 'HR Directory',
    '{"name":"HR Directory","details":"Full employee directory for HR","columns":["EMPLOYEE_ID","FIRST_NAME","LAST_NAME","EMAIL","PHONE","<SPACE>","<COMMENTARY> Employment","DEPARTMENT","JOB_TITLE","SALARY","HIRE_DATE","IS_ACTIVE","<SPACE>","<COMMENTARY> Reporting","MANAGER_ID"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'hr_team', 'EMPLOYEES', 'By Department',
    '{"name":"By Department","query":"SELECT * FROM {THIS_TABLE} WHERE DEPARTMENT = :DEPT ORDER BY LAST_NAME","arguments":[{"name":"DEPT","type":"string","description":"Department name (Engineering, Sales, HR, Finance)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'hr_team', 'EMPLOYEES', 'New Hires (After Date)',
    '{"name":"New Hires (After Date)","query":"SELECT * FROM {THIS_TABLE} WHERE HIRE_DATE >= TO_DATE(:AFTER_DATE, ''YYYY-MM-DD'') ORDER BY HIRE_DATE DESC","arguments":[{"name":"AFTER_DATE","type":"string","description":"Date (YYYY-MM-DD)"}]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'hr_team', 'EMPLOYEES', 'Inactive Employees',
    '{"name":"Inactive Employees","query":"SELECT * FROM {THIS_TABLE} WHERE IS_ACTIVE = ''N''","arguments":[]}'
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Client: hr_team / Table: AUDIT_LOG
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_FILTER', 'hr_team', 'AUDIT_LOG', 'Employee Changes',
    '{"name":"Employee Changes","details":"Audit trail for employee records","columns":["LOG_ID","TABLE_NAME","OPERATION","RECORD_ID","CHANGED_BY","CHANGED_AT","<THE REST>"]}'
);
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'PRESET_QUERY', 'hr_team', 'AUDIT_LOG', 'Employee Audits',
    '{"name":"Employee Audits","query":"SELECT * FROM {THIS_TABLE} WHERE TABLE_NAME = ''EMPLOYEES'' ORDER BY CHANGED_AT DESC","arguments":[]}'
);

-- =============================================================================
-- CLIENT_CONFIG — Seed client connection configs from config.yml
-- =============================================================================

-- Client: local_test
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'CLIENT_CONFIG', NULL, NULL, 'local_test',
    '{"name":"local_test","display_name":"Local Test DB","host":"localhost","port":1521,"service_name":"XE","username":"LEARN1","password":"Welcome","schema":"LEARN1","tables":["<COMMENTARY> CRM","CUSTOMERS","PRODUCTS","<SPACE>","<COMMENTARY> Sales","ORDERS","<SPACE>","<COMMENTARY> HR","EMPLOYEES","<SPACE>","<COMMENTARY> Reference","CATEGORIES","<SPACE>","<COMMENTARY> System","AUDIT_LOG","DOCUMENTS"]}'
);

-- Client: sales_team
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'CLIENT_CONFIG', NULL, NULL, 'sales_team',
    '{"name":"sales_team","display_name":"Sales Team","host":"localhost","port":1521,"service_name":"XE","username":"LEARN1","password":"Welcome","schema":"LEARN1","tables":["CUSTOMERS","ORDERS","PRODUCTS"]}'
);

-- Client: hr_team
INSERT INTO DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA) VALUES (
    'CLIENT_CONFIG', NULL, NULL, 'hr_team',
    '{"name":"hr_team","display_name":"HR Team","host":"localhost","port":1521,"service_name":"XE","username":"LEARN1","password":"Welcome","schema":"LEARN1","tables":["EMPLOYEES","AUDIT_LOG"]}'
);

COMMIT;
PROMPT === DB_VIEWER_APP_DATA setup complete ===
PROMPT Table created: DB_VIEWER_APP_DATA (with 2 indexes)
PROMPT Seed data: preset filters and queries for clients local_test, sales_team, hr_team
PROMPT Seed data: CLIENT_CONFIG for 3 clients (local_test, sales_team, hr_team)
