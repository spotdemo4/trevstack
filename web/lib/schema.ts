import {
	create,
	type DescMessage,
	type MessageInitShape,
} from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";

export function createSchema<Desc extends DescMessage>(messageDesc: Desc) {
	const schema = createStandardSchema(messageDesc);

	const wrappedValidate = (value: MessageInitShape<Desc>) =>
		schema["~standard"].validate(create(messageDesc, value));

	return {
		...schema,
		"~standard": {
			...schema["~standard"],
			validate: wrappedValidate,
		},
	} as typeof schema;
}
