import { useState } from "react";

interface InputBoxProps {
  onSubmit: (product: string) => void;
  loading: boolean;
}

export default function InputBox({ onSubmit, loading }: InputBoxProps) {
  const [product, setProduct] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!product.trim()) return;
    onSubmit(product);
    setProduct("");
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={product}
        onChange={(e) => setProduct(e.target.value)}
        placeholder="Enter a product URL or description"
        disabled={loading}
      />
      <button type="submit" disabled={loading}>
        Submit
      </button>
    </form>
  );
}
