# Model
Data model for items, references and sub items.

## Items
An item is a struct that may exist on its own to store the fields in the struct.
Some of the fields may be references to other items
List fields must be stored as sub items

## Sub Items
A sub item exist as part of its parent item. When the parent is deleted, the sub item is also deleted.
It may also store data fields and references like an item.

## References
A reference is a field inside an item or sub-item that refers to another item.
The other item is stored separately and the reference only stores the ID of that item.

## Conclusion
An item has a unique ID
A sub item has a parent ID and other fields that makes it unique within the parent# model
