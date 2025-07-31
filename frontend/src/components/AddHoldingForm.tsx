import { useForm } from 'react-hook-form';

import { zodResolver } from '@hookform/resolvers/zod';

import { z } from 'zod';

import { useMutation, useQueryClient } from 'react-query';

import { useEffect, useState } from 'react';

import { QUERY_KEYS } from '@/types/portfolio';

import { Button, Card, LoadingSpinner } from './ui';

import { UI_CONSTANTS } from './ui';

import AssetSelector from './AssetSelector';

import { AssetSearchResult, TrendingAsset } from '@/hooks/useMarketData';

  

const holdingSchema = z.object({

symbol: z.string().min(1, 'Symbol is required'),

quantity: z.preprocess(

(val) => Number(val),

z.number({

required_error: 'Quantity is required',

invalid_type_error: 'Quantity must be a number'

}).positive('Quantity must be positive')

),

average_cost: z.preprocess(

(val) => Number(val),

z.number({

required_error: 'Average cost is required',

invalid_type_error: 'Average cost must be a number'

}).positive('Average cost must be positive')

),

});

  

type HoldingFormData = z.infer<typeof holdingSchema>;

  

interface AddHoldingFormProps {

preSelectedAsset?: {

symbol: string;

name: string;

} | null;

}

  

export default function AddHoldingForm({ preSelectedAsset }: AddHoldingFormProps) {

const queryClient = useQueryClient();

const [selectedAssetInfo, setSelectedAssetInfo] = useState<AssetSearchResult | TrendingAsset | null>(null);

  

const {

register,

handleSubmit,

formState: { errors },

reset,

watch,

setValue,

control,

} = useForm<HoldingFormData>({

resolver: zodResolver(holdingSchema),

defaultValues: {

symbol: preSelectedAsset?.symbol || '',

quantity: undefined,

average_cost: undefined

}

});

  

const formValues = watch();

  

// Update form when preSelectedAsset changes

useEffect(() => {

if (preSelectedAsset) {

reset({

symbol: preSelectedAsset.symbol,

quantity: undefined,

average_cost: undefined

});

}

}, [preSelectedAsset, reset]);

  

const handleAssetSelect = (asset: AssetSearchResult | TrendingAsset) => {

setSelectedAssetInfo(asset);

setValue('symbol', asset.symbol);

};

  

const handleSymbolChange = (symbol: string) => {

setValue('symbol', symbol);

// Clear selected asset info if user types manually

if (selectedAssetInfo && selectedAssetInfo.symbol !== symbol) {

setSelectedAssetInfo(null);

}

};

  

const {

mutate,

isLoading,

error: mutationError

} = useMutation({

  

mutationFn: async (data: HoldingFormData) => {

const response = await fetch('/api/v1/portfolio/holdings', {

method: 'POST',

headers: {

'Content-Type': 'application/json',

},

body: JSON.stringify(data),

});

if (!response.ok) {

const error = await response.json();

console.error('API Error:', error);

throw new Error(error.message || 'Failed to add holding');

}

return response.json();

},

onSuccess: () => {

queryClient.invalidateQueries([QUERY_KEYS.PORTFOLIO_HOLDINGS]);

queryClient.invalidateQueries([QUERY_KEYS.PORTFOLIO_SUMMARY]);

// Reset form values while maintaining proper types

reset();

// reset({ symbol: '', quantity: null, average_cost: null});

// reset({

// symbol: '',

// quantity: undefined,

// average_cost: undefined

// });

// Force empty string values for number inputs to clear them visually

setTimeout(() => {

const quantityInput = document.getElementById('quantity') as HTMLInputElement;

const costInput = document.getElementById('average_cost') as HTMLInputElement;

if (quantityInput) quantityInput.value = '';

if (costInput) costInput.value = '';

}, 0);

},

onError: (error: Error) => {

// Error will be displayed in the UI

},

});

  

const onSubmit = (data: HoldingFormData) => {

mutate(data);

};

  

return (

<Card title="Add New Holding" className={UI_CONSTANTS.spacing.section}>

<form onSubmit={handleSubmit(onSubmit)} className="space-y-4">

<div>

<label htmlFor="symbol">Symbol</label>

<AssetSelector

id="symbol"

value={formValues.symbol}

onChange={handleSymbolChange}

onAssetSelect={handleAssetSelect}

placeholder="Search for stocks, ETFs, crypto..."

error={errors.symbol?.message}

showTrending={true}

enableDetailsModal={true}

/>

{/* Hidden input for form validation */}

<input

type="hidden"

{...register('symbol')}

value={formValues.symbol}

/>

{selectedAssetInfo && (

<div className="mt-2 p-2 bg-blue-50 rounded-md">

<p className="text-sm text-blue-800">

<strong>{selectedAssetInfo.symbol}</strong> - {selectedAssetInfo.name}

{selectedAssetInfo.exchange && (

<span className="text-blue-600 ml-2">({selectedAssetInfo.exchange})</span>

)}

</p>

{'current_price' in selectedAssetInfo && (

<p className="text-xs text-blue-600 mt-1">

Current Price: ${selectedAssetInfo.current_price?.toFixed(2)}

</p>

)}

</div>

)}

</div>

  

<div>

<label htmlFor="quantity">Quantity</label>

<input

id="quantity"

type="number"

step="any"

{...register('quantity', {

valueAsNumber: true,

setValueAs: (v) => v === '' ? undefined : Number(v),

})}

className="w-full p-2 border rounded"

/>

{errors.quantity && (

<p className="text-red-500 text-sm">{errors.quantity.message}</p>

)}

</div>

  

<div>

<label htmlFor="average_cost">Average Cost</label>

<input

id="average_cost"

type="number"

step="any"

{...register('average_cost', {

valueAsNumber: true,

setValueAs: (v) => v === '' ? undefined : Number(v),

})}

className="w-full p-2 border rounded"

/>

{errors.average_cost && (

<p className="text-red-500 text-sm">{errors.average_cost.message}</p>

)}

</div>

  

<div className="flex justify-end space-x-2 pt-4">

<button

type="button"

onClick={() => reset()}

className="px-4 py-2 bg-gray-200 rounded"

>

Reset

</button>

<button

type="submit"

disabled={isLoading}

className="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"

>

{isLoading ? (

<>

<LoadingSpinner size="sm" className="mr-2" />

Adding...

</>

) : (

'Add Holding'

)}

</button>

</div>

{mutationError && (

<p className="text-red-500 text-sm mt-2">

{mutationError.message || 'Failed to add holding'}

</p>

)}

</form>

</Card>

);

}