import {
	create,
	type DescMessage,
	type MessageInitShape,
} from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";

export function createSchema<Desc extends DescMessage>(messageDesc: Desc) {
	const schema = createStandardSchema(messageDesc);

	const wrappedValidate = async (value: MessageInitShape<Desc>) => {
		const result = schema["~standard"].validate(create(messageDesc, value));
		return Promise.resolve(result);
	};

	return {
		...schema,
		"~standard": {
			...schema["~standard"],
			validate: wrappedValidate,
		},
	} as typeof schema;
}
