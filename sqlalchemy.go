package model

//database client interface
type IClient interface {
	GetSessionContext(name string) ISession //name="stock.leader"
}

type ISession interface {
	Query(items ...IItem /*e.g. StockLevel, StockStatus*/) IResults
}

type IResults interface {
	Join(IItem) IResults
	Filter(expressions ...IExpression) IResults
	Merged() IResults
	All() []interface{}
}

type IExpression interface{}

/*
 * examples from python:

 import random
import time

from model import Column, ForeignKey, UniqueConstraint
from model.dialects import mysql
from model.exc import DataError, IntegrityError, OperationalError
from sql_client import SQLClient

from stock_service.models.take2 import Base
from stock_service.models.base_model import BaseModel
from stock_service.models.inventory import Inventory, InventoryModel
from stock_service.models.location import Location, LocationModel
from stock_service.models.owner import Owner
from stock_service.models.stock import StockModel, Stock
from stock_service.models.stock_status import StockStatusModel, StockStatus

DEFAULT_DEADLOCK_RETRY_WINDOW_SECONDS = 5


class StockLevel(Base):
    """ Stock level table """
    __tablename__ = "stock_level"
    __table_args__ = (
        UniqueConstraint('stock_id', 'stock_status_id', name='uniqueStockLevel'),
    )

    stock_id = Column("stock_id", mysql.INTEGER(11), ForeignKey("stock.stock_id"), primary_key=True)
    stock_status_id = Column("stock_status_id", mysql.INTEGER(11), ForeignKey("stock_status.stock_status_id"),
                             primary_key=True)
    quantity = Column("quantity", mysql.INTEGER(11))
    date_created = Column("date_created", mysql.DATETIME())
    date_modified = Column("date_modified", mysql.DATETIME())


class StockLevelModel(BaseModel):
    """ Model for accessing the stock level table """
    def __init__(self, client: SQLClient, deadlock_retry_window=DEFAULT_DEADLOCK_RETRY_WINDOW_SECONDS):
        super().__init__(client, StockLevel)
        self.inventory_model = InventoryModel(client)
        self.location_model = LocationModel(client)
        self.stock_model = StockModel(client)
        self.stock_status_model = StockStatusModel(client)
        self.deadlock_retry_window = deadlock_retry_window

    def get_by(self, stock_id=None, stock_status_id=None):
        kwargs = {k: v for k, v in {"stock_id": stock_id, "stock_status_id": stock_status_id}.items() if v is not None}
        return super().get_all(**kwargs)

    def create(self, stock_id, stock_status_id, quantity):
        try:
            return super().create(stock_id=stock_id, stock_status_id=stock_status_id, quantity=quantity)
        except (IntegrityError, DataError) as e:
            if not self.stock_model.get_by_id(stock_id=stock_id):
                raise ReferenceError(f"Cannot find stock id {stock_id}") from e
            if not self.stock_status_model.get_by_id(stock_status_id=stock_status_id):
                raise ReferenceError(f"Cannot find stock status id {stock_status_id}") from e
            if self.get_by(stock_id=stock_id, stock_status_id=stock_status_id):
                raise ValueError(f"Duplicate entry: stock level already exists - stock: {stock_id}, "
                                 f"stock status: {stock_status_id}") from e
            raise

    def get_level(self, product_reference, company_id):
        with self.client.get_session_context("stock.leader") as session:
            results = session.query(StockLevel).join(Stock).join(Inventory).join(Owner).\
                filter(Owner.company_id == company_id,
                       Inventory.product_reference == product_reference).all()
            return [r.to_dict() for r in results]

    def get_level_by_location(self, product_reference, company_id, location_name):
        with self.client.get_session_context("stock.leader") as session:
            results = session.query(StockLevel).join(Stock).join(Inventory).join(Owner).join(Location).\
                filter(Owner.company_id == company_id,
                       Inventory.product_reference == product_reference,
                       Location.name == location_name,
                       Location.location_id == Stock.location_id).all()
            return [r.to_dict() for r in results]

    def get_levels_by_location_id(self, product_reference, company_id, location_id):
        with self.client.get_session_context("stock.leader") as session:
            results = session.query(StockLevel, StockStatus).join(Stock).join(Inventory).join(Owner).join(StockStatus).\
                filter(Owner.company_id == company_id,
                       Inventory.product_reference == product_reference,
                       StockLevel.stock_status_id == StockStatus.stock_status_id,
                       Stock.location_id == location_id).all()
            return [{**r.StockLevel.to_dict(), **r.StockStatus.to_dict()} for r in results]

    def get_level_by_status(self, product_reference, company_id, stock_status_id):
        with self.client.get_session_context("stock.leader") as session:
            results = session.query(StockLevel).join(Stock).join(Inventory).join(Owner).\
                filter(Owner.company_id == company_id,
                       Inventory.product_reference == product_reference,
                       StockLevel.stock_status_id == stock_status_id).all()
            return [r.to_dict() for r in results]

    def get_level_by_location_and_status(self, product_reference, company_id, location_name, stock_status_id):
        with self.client.get_session_context("stock.leader") as session:
            results = session.query(StockLevel).join(Stock).join(Inventory).join(Owner).join(Location).\
                filter(Owner.company_id == company_id,
                       Inventory.product_reference == product_reference,
                       Location.name == location_name,
                       Location.location_id == Stock.location_id,
                       StockLevel.stock_status_id == stock_status_id).all()
            return [r.to_dict() for r in results]

    def create_level(self, product_reference, company_id, location_name, stock_status_id, quantity):
        inventory = self.inventory_model.get_by_product_and_company(product_reference=product_reference, company_id=company_id)
        if not inventory:
            raise ReferenceError(f'Found no inventory for: {{"product": {product_reference}, "company": {company_id}}}')
        location = self.location_model.get_by(name=location_name)
        if not location:
            raise ReferenceError(f'Found no location for: {{"location": {location_name}}}')
        if len(location) > 1:
            raise ReferenceError(f'Found too many locations [{len(location)} > 1] for: {{"location": {location_name}}}')
        inventory_id = inventory[0]["inventory_id"]
        location_id = location[0]["location_id"]
        stock = self.stock_model.get_by(inventory_id=inventory_id, location_id=location_id)
        if not stock:
            raise ReferenceError(f'Found no stock for: '
                                 f'{{"inventory_id": {inventory_id}, "location_id": {location_id}}}')
        if len(stock) > 1:
            raise ReferenceError(f'Found too much stock [{len(stock)} > 1] for:'
                                 f'{{"inventory_id": {inventory_id}, "location_id": {location_id}}}')
        self.create(stock[0]["stock_id"], stock_status_id=stock_status_id, quantity=quantity)

    def set_level(self, product_reference, company_id, location_name, stock_status_id, quantity):
        # self._validate_stock_modification(quantity)
        try:
            with self.client.get_session_context("stock.leader") as session:
                stock_level = self._query_stock_levels(session, company_id, product_reference, location_name,
                                                       [stock_status_id])[0]
                stock_level.quantity = quantity
                session.merge(stock_level)
                session.commit()
        except ReferenceError:
            # if we haven't found a stock level, then we must create one
            self.create_level(product_reference, company_id, location_name, stock_status_id, quantity)

    def increase_level(self, product_reference, company_id, location_name, stock_status_id, quantity):
        # self._validate_stock_modification(quantity)
        with self.client.get_session_context("stock.leader") as session:
            stock_level = self._query_stock_levels(session, company_id, product_reference, location_name, [stock_status_id])[0]
            stock_level.quantity += quantity
            session.merge(stock_level)
            session.commit()

    def decrease_level(self, product_reference, company_id, location_name, stock_status_id, quantity):
        # self._validate_stock_modification(quantity)
        with self.client.get_session_context("stock.leader") as session:
            stock_level = self._query_stock_levels(session, company_id, product_reference, location_name, [stock_status_id])[0]
            if quantity > stock_level.quantity:
                raise ValueError(f"Stock quantity cannot go negative (current stock: {stock_level.quantity}, "
                                 f"decrease adjustment: {quantity})")
            stock_level.quantity -= quantity
            session.merge(stock_level)
            session.commit()

    def move_between_levels(self, product_reference, company_id, location_name, original_stock_status_id,
                            final_stock_status_id, quantity, retry_for_missing_stock_level=True):
        start = time.time()
        while True:
            with self.client.get_session_context("stock.leader") as session:
                try:
                    levels = self._query_stock_levels(session, company_id, product_reference, location_name,
                                                      [original_stock_status_id, final_stock_status_id])
                    from_level = levels[0]
                    to_level = levels[1]
                    if quantity > from_level.quantity:
                        raise ValueError(f"Stock quantity cannot go negative (current stock: {from_level.quantity}, "
                                         f"decrease adjustment: {quantity})")
                    from_level.quantity -= quantity
                    to_level.quantity += quantity
                    session.merge(from_level)
                    session.merge(to_level)
                    session.commit()
                    return
                except OperationalError as e:
                    time.sleep(random.random() / 100)
                    if time.time() - start > self.deadlock_retry_window:
                        raise RuntimeError(f'Deadlock found in DB lookup for: {{"product_reference": {product_reference}, '
                                           f'"company_id": {company_id}, "location_name": \"{location_name}\", '
                                           f'"stock_statuses": [{original_stock_status_id}, {final_stock_status_id}]}}')
                except ReferenceError:
                    # If we are unable to find both levels then assume that the final stock status does not exist,
                    #   create it and try again
                    if retry_for_missing_stock_level:
                        retry_for_missing_stock_level = False
                        break
                    # Otherwise re-raise because we are missing the original stock status
                    raise
        self.create_level(product_reference, company_id, location_name, final_stock_status_id, 0)
        return self.move_between_levels(product_reference, company_id, location_name, original_stock_status_id,
                                        final_stock_status_id, quantity, retry_for_missing_stock_level)

    # Utility methods
    @staticmethod
    def _query_stock_levels(session, company_id, product_reference, location_name, stock_status_id_list):
        response = session.query(StockLevel).with_for_update().join(Stock).join(Inventory).join(Owner).join(Location).\
            filter(Owner.company_id == company_id,
                   Inventory.product_reference == product_reference,
                   Location.name == location_name,
                   Location.location_id == Stock.location_id,
                   StockLevel.stock_status_id.in_(stock_status_id_list)).all()
        ordered_levels = []
        for stock_status_id in stock_status_id_list:
            level = [level for level in response if level.stock_status_id == stock_status_id]
            if not level:
                raise ReferenceError(f'Cannot find stock level for: {{"product_reference": {product_reference}, '
                                     f'"company_id": {company_id}, "location_name": \"{location_name}\", '
                                     f'"stock_status_id": {stock_status_id}}}')
            ordered_levels.append(level.pop())
        return ordered_levels


*/
