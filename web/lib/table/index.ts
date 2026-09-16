import { Body, Header, Table as TableRoot } from "./table";

export type { TableScrollMode } from "./table";
export { Body, Header, TableRoot };

export const Table = {
  Table: TableRoot,
  Header,
  Body,
};
