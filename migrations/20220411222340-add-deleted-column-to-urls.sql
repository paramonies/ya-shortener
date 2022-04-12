-- +migrate Up
alter table urls add column deleted bool default false;
-- +migrate Down
alter table urls drop column deleted;