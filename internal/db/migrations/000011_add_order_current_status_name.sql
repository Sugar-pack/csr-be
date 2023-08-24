-- +migrate Up
ALTER TABLE "orders" ADD COLUMN "order_status_name_orders" INTEGER NULL;
ALTER TABLE "orders" ADD CONSTRAINT "orders_order_status_name_orders_fkey" FOREIGN KEY("order_status_name_orders") REFERENCES "order_status_names"("id") ON DELETE SET NULL;

UPDATE orders o
SET order_status_name_orders = (
    SELECT os.order_status_name_order_status
    FROM order_status os
    WHERE os.order_order_status = o.id
    ORDER BY os.current_date DESC
    LIMIT 1
)
WHERE o.id IN (SELECT DISTINCT order_order_status FROM order_status);

-- +migrate Down
ALTER TABLE "orders" DROP COLUMN "order_status_name_orders";
